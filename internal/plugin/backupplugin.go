/*
Copyright 2017, 2019 the Velero contributors.

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
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
)

const (
	// If this annotation is found on the Velero Backup CR, then sleep
	// for the specified duration (logging before and after)
	// This will facilitate testing of parallel item backup functionality
	// This annotation can also be set on the item, which overrides the backup CR value
	BIAWaitDurationAnnotation = "velero.io/example-bia-wait-duration"
)

// BackupPlugin is a backup item action plugin for Velero.
type BackupPlugin struct {
	log logrus.FieldLogger
}

// NewBackupPlugin instantiates a BackupPlugin.
func NewBackupPlugin(log logrus.FieldLogger) *BackupPlugin {
	return &BackupPlugin{log: log}
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A BackupPlugin's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *BackupPlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{}, nil
}

// Execute allows the ItemAction to perform arbitrary logic with the item being backed up,
// in this case, setting a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	p.log.Info("Hello from my BackupPlugin(v1)!")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["velero.io/my-backup-plugin"] = "1"

	metadata.SetAnnotations(annotations)

	var duration time.Duration
	if durationStr, ok := annotations[BIAWaitDurationAnnotation]; ok && len(durationStr) != 0 {
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			p.log.Warnf("Error parsing duration on item: %v", err)
		}
	}
	if duration == 0 && backup.Annotations != nil {
		if durationStr, ok := backup.Annotations[BIAWaitDurationAnnotation]; ok && len(durationStr) != 0 {
			duration, err = time.ParseDuration(durationStr)
			if err != nil {
				p.log.Warnf("Error parsing duration on Backup: %v", err)
			}
		}
	}

	if duration != 0 {
		p.log.Infof("BIA for %v, %v/%v, waiting %v", item.GetObjectKind().GroupVersionKind().Kind, metadata.GetNamespace(), metadata.GetName(), duration)
		time.Sleep(duration)
		p.log.Infof("BIA for %v, %v/%v, done waiting", item.GetObjectKind().GroupVersionKind().Kind, metadata.GetNamespace(), metadata.GetName())
	}
	return item, nil, nil
}
