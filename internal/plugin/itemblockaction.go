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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/client"
	"github.com/vmware-tanzu/velero/pkg/kuberesource"
	plugincommon "github.com/vmware-tanzu/velero/pkg/plugin/framework/common"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
)

// ItemBlockActionPlugin implements ItemBlockAction.
type ItemBlockActionPlugin struct {
	log logrus.FieldLogger
	crClient crclient.Client
}

// NewItemBlockActionPlugin creates a new ItemBlockAction for pods.
func NewItemBlockActionPlugin(f client.Factory) plugincommon.HandlerInitializer {
	return func(logger logrus.FieldLogger) (interface{}, error) {
		crClient, err := f.KubebuilderClient()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return &ItemBlockActionPlugin{
			log:      logger,
			crClient: crClient,
		}, nil
	}
}

// AppliesTo returns a ResourceSelector that applies only to pods.
func (a *ItemBlockActionPlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"pods"},
	}, nil
}

// GetRelatedItems executes arbitrary logic for individual items to determine which items should be backed up together
// In this case, it looks for other pods in the same namespace with the same `itemblock` label and returns those
// as items to back up with this one.
func (a *ItemBlockActionPlugin) GetRelatedItems(item runtime.Unstructured, backup *v1.Backup) ([]velero.ResourceIdentifier, error) {
	a.log.Info("Executing example pod ItemBlockAction")
	defer a.log.Info("Done executing example pod ItemBlockAction")

	pod := new(corev1api.Pod)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.UnstructuredContent(), pod); err != nil {
		return nil, errors.WithStack(err)
	}
	return a.relatedItemsForPod(pod, a.log)
}

func (a *ItemBlockActionPlugin) relatedItemsForPod(pod *corev1api.Pod, log logrus.FieldLogger) ([]velero.ResourceIdentifier, error) {
	labels := pod.ObjectMeta.Labels
	var itemblockLabel string
	var relatedItems []velero.ResourceIdentifier

	if labels != nil {
		itemblockLabel = labels["itemblock"]
	}
	if len(itemblockLabel) > 0 {
		pods := new(corev1api.PodList)
		err := a.crClient.List(context.Background(), pods, crclient.InNamespace(pod.Namespace), crclient.MatchingLabels{"itemblock": itemblockLabel})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list pods")
		}
		for i := range pods.Items {
			if pods.Items[i].Name != pod.Name {
				relatedItems = append(relatedItems, velero.ResourceIdentifier{
					GroupResource: kuberesource.Pods,
					Namespace:     pods.Items[i].Namespace,
					Name:          pods.Items[i].Name,
				})
			}
		}
	}

	return relatedItems, nil
}

func (a *ItemBlockActionPlugin) Name() string {
	return "PodItemBlockAction"
}
