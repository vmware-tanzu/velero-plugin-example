/*
Copyright the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	corev1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/pkg/errors"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/kuberesource"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	biav2 "github.com/vmware-tanzu/velero/pkg/plugin/velero/backupitemaction/v2"
)

const (
	// If this annotation is found on the Velero Backup CR, then create an operation
	// that is considered done at backup start time + example BIA operation duration
	// If this annotation is not present, then operationID returned from Execute() will
	// be empty.
	// This annotation can also be set on the item, which overrides the backup CR value,
	// to allow for testing multiple action lengths
	AsyncBIADurationAnnotation = "velero.io/example-bia-operation-duration"
	// If this annotation is true on the item, then if the BIA duration is set to a
	// non-zero value, a Secret will be created, and it will be returned as an additional
	// item with the UpdateAdditionalItemsAfterOperation return flag set to true
	AsyncBIAAdditionalUpdateAnnotation = "velero.io/example-bia-additional-update"
	AsyncBIAProgressAnnotation         = "velero.io/example-bia-progress"
	AsyncBIAExampleSecretAnnotation    = "velero.io/example-bia-secret"
	AsyncBIAExampleLabel               = "velero.io/example-bia"
)

// BackupPluginV2 is a v2 backup item action plugin for Velero.
type BackupPluginV2 struct {
	log logrus.FieldLogger
}

// NewBackupPluginV2 instantiates a v2 BackupPlugin.
func NewBackupPluginV2(log logrus.FieldLogger) *BackupPluginV2 {
	return &BackupPluginV2{log: log}
}

// Name is required to implement the interface, but the Velero pod does not delegate this
// method -- it's used to tell velero what name it was registered under. The plugin implementation
// must define it, but it will never actually be called.
func (p *BackupPluginV2) Name() string {
	return "exampleBackupPlugin"
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A BackupPlugin's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *BackupPluginV2) AppliesTo() (velero.ResourceSelector, error) {
	// exclude secrets to avoid infinite loop, since this plugin creates secrets as additional items.
	return velero.ResourceSelector{ExcludedResources: []string{"secrets"}}, nil
}

func GetClient() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return client, nil
}

// Execute allows the ItemAction to perform arbitrary logic with the item being backed up,
// in this case, setting a custom annotation on the item being backed up.
func (p *BackupPluginV2) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, string, []velero.ResourceIdentifier, error) {
	p.log.Info("Hello from my BackupPlugin(v2)!")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, "", nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["velero.io/my-backup-pluginv2"] = "1"

	metadata.SetAnnotations(annotations)

	// Operations during finalize aren't supported, so if backup is in a finalize phase, just return the item
	if backup.Status.Phase == v1.BackupPhaseFinalizing ||
		backup.Status.Phase == v1.BackupPhaseFinalizingPartiallyFailed {
		return item, nil, "", nil, nil
	}

	operationID := ""
	duration := ""
	if durationStr, ok := annotations[AsyncBIADurationAnnotation]; ok && len(durationStr) != 0 {
		_, err := time.ParseDuration(durationStr)
		if err == nil {
			duration = durationStr
		}
	}
	if duration == "" && backup.Annotations != nil {
		if durationStr, ok := backup.Annotations[AsyncBIADurationAnnotation]; ok && len(durationStr) != 0 {
			_, err := time.ParseDuration(durationStr)
			if err == nil {
				duration = durationStr
			}
		}
	}
	// If duration is empty, we don't have an operation so just return the item.
	if duration == "" {
		return item, nil, "", nil, nil
	}

	var secret *corev1api.Secret
	var itemsToUpdate []velero.ResourceIdentifier
	additionalUpdate, ok := annotations[AsyncBIAAdditionalUpdateAnnotation]
	if ok && additionalUpdate == "true" {
		secret = &corev1api.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    metadata.GetNamespace(),
				GenerateName: metadata.GetName() + "-",
				Labels: map[string]string{
					AsyncBIAExampleLabel: "true",
				},
			},
			Type: corev1api.SecretTypeOpaque,
			Data: map[string][]byte{
				"TestObject": []byte(metadata.GetName()),
			},
		}
		secretClient, err := GetClient()
		if err != nil {
			return item, nil, "", nil, errors.Wrap(err, "error getting secret client")
		}
		if secret, err = secretClient.CoreV1().Secrets(metadata.GetNamespace()).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
			return item, nil, "", nil, errors.Wrapf(err, "error creating %s secret", metadata.GetName())
		}
	}

	operationID = string(metadata.GetUID()) + "/" + duration
	if secret != nil {
		itemsToUpdate = []velero.ResourceIdentifier{
			{
				GroupResource: kuberesource.Secrets,
				Namespace:     secret.Namespace,
				Name:          secret.Name,
			},
		}
		operationID += "/" + secret.Namespace + "/" + secret.Name
		annotations[AsyncBIAExampleSecretAnnotation] = secret.Name
		metadata.SetAnnotations(annotations)

	}
	return item, nil, operationID, itemsToUpdate, nil
}

func (p *BackupPluginV2) Progress(operationID string, backup *v1.Backup) (velero.OperationProgress, error) {
	progress := velero.OperationProgress{}
	if operationID == "" {
		return progress, biav2.InvalidOperationIDError(operationID)
	}
	splitOp := strings.Split(operationID, "/")
	if len(splitOp) == 4 {
		secretClient, err := GetClient()
		if err != nil {
			return progress, errors.Wrap(err, "error getting secret client")
		}
		secret, err := secretClient.CoreV1().Secrets(splitOp[2]).Get(context.TODO(), splitOp[3], metav1.GetOptions{})
		if err != nil {
			return progress, errors.Wrapf(err, "error getting %s secret", splitOp[3])
		}
		annotations := secret.Annotations
		if annotations == nil {
			annotations = make(map[string]string)
		}
		priorProgressCalls := 0
		progressAnnotation, ok := annotations[AsyncBIAProgressAnnotation]
		if ok {
			i, err := strconv.Atoi(progressAnnotation)
			if err == nil {
				priorProgressCalls = i
			}
		}
		annotations[AsyncBIAProgressAnnotation] = strconv.Itoa(priorProgressCalls + 1)
		secret.Annotations = annotations
		if _, err := secretClient.CoreV1().Secrets(splitOp[2]).Update(context.TODO(), secret, metav1.UpdateOptions{}); err != nil {
			return progress, errors.Wrapf(err, "error updating %s secret", splitOp[3])
		}

	} else if len(splitOp) != 2 {
		return progress, biav2.InvalidOperationIDError(operationID)
	}
	duration, err := time.ParseDuration(splitOp[1])
	if err != nil {
		return progress, biav2.InvalidOperationIDError(operationID)
	}
	elapsed := time.Since(backup.Status.StartTimestamp.Time).Seconds()
	if elapsed >= duration.Seconds() {
		progress.Completed = true
		progress.NCompleted = int64(duration.Seconds())
	} else {
		progress.NCompleted = int64(elapsed)
	}
	progress.NTotal = int64(duration.Seconds())
	progress.OperationUnits = "seconds"
	progress.Updated = time.Now()

	return progress, nil
}

func (p *BackupPluginV2) Cancel(operationID string, backup *v1.Backup) error {
	return nil
}
