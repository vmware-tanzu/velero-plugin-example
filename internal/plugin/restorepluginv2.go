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
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	riav2 "github.com/vmware-tanzu/velero/pkg/plugin/velero/restoreitemaction/v2"
	"k8s.io/apimachinery/pkg/api/meta"
)

const (
	// If this annotation is found on the Velero Restore CR, then create an operation
	// that is considered done at backup start time + example RIA operation duration
	// If this annotation is not present, then operationID returned from Execute() will
	// be empty.
	// This annotation can also be set on the item, which overrides the restore CR value,
	// to allow for testing multiple action lengths
	AsyncRIADurationAnnotation         = "velero.io/example-ria-operation-duration"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePluginV2 struct {
	log logrus.FieldLogger
}

// NewRestorePluginV2 instantiates a v2 RestorePlugin.
func NewRestorePluginV2(log logrus.FieldLogger) *RestorePluginV2 {
	return &RestorePluginV2{log: log}
}

// Name is required to implement the interface, but the Velero pod does not delegate this
// method -- it's used to tell velero what name it was registered under. The plugin implementation
// must define it, but it will never actually be called.
func (p *RestorePluginV2) Name() string {
	return "exampleRestorePlugin"
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A RestoreItemAction's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *RestorePluginV2) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{}, nil
}

// Execute allows the RestorePlugin to perform arbitrary logic with the item being restored,
// in this case, setting a custom annotation on the item being restored.
func (p *RestorePluginV2) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.log.Info("Hello from my RestorePlugin(v2)!")

	metadata, err := meta.Accessor(input.Item)
	if err != nil {
		return &velero.RestoreItemActionExecuteOutput{}, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["velero.io/my-restore-pluginv2"] = "1"

	metadata.SetAnnotations(annotations)

	duration := ""
	if durationStr, ok := annotations[AsyncRIADurationAnnotation]; ok && len(durationStr) != 0 {
		_, err := time.ParseDuration(durationStr)
		if err == nil {
			duration = durationStr
		}
	}
	if duration == "" && input.Restore.Annotations != nil {
		if durationStr, ok := input.Restore.Annotations[AsyncRIADurationAnnotation]; ok && len(durationStr) != 0 {
			_, err := time.ParseDuration(durationStr)
			if err == nil {
				duration = durationStr
			}
		}
	}
	out := velero.NewRestoreItemActionExecuteOutput(input.Item)
	// If duration is empty, we don't have an operation so just return the item.
	if duration != "" {
		out = out.WithOperationID(string(metadata.GetName()) + "/" + duration)
	}

	return out, nil
}

func (p *RestorePluginV2) Progress(operationID string, restore *v1.Restore) (velero.OperationProgress, error) {
	progress := velero.OperationProgress{}
	if operationID == "" {
		return progress, riav2.InvalidOperationIDError(operationID)
	}
	splitOp := strings.Split(operationID, "/")
	if len(splitOp) != 2 {
		return progress, riav2.InvalidOperationIDError(operationID)
	}
	duration, err := time.ParseDuration(splitOp[1])
	if err != nil {
		return progress, riav2.InvalidOperationIDError(operationID)
	}
	elapsed := time.Since(restore.Status.StartTimestamp.Time).Seconds()
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

func (p *RestorePluginV2) Cancel(operationID string, restore *v1.Restore) error {
	return nil
}

func (p *RestorePluginV2) AreAdditionalItemsReady(additionalItems []velero.ResourceIdentifier, restore *v1.Restore) (bool, error) {
	return true, nil
}
