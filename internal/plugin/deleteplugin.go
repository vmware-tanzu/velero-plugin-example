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
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/api/meta"
)

// DeletePlugin is a delete item action plugin for Velero
type DeletePlugin struct {
	log logrus.FieldLogger
}

// NewDeletePlugin instantiates a DeletePlugin.
func NewDeletePlugin(log logrus.FieldLogger) *DeletePlugin {
	return &DeletePlugin{log: log}
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A DeleteItemAction's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *DeletePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{}, nil
}

// Execute allows the DeletePlugin to perform arbitrary logic with the item being deleted,
func (p *DeletePlugin) Execute(input *velero.DeleteItemActionExecuteInput) error {
	p.log.Info("Hello from my DeletePlugin!")
	defer p.log.Info("Done executing my DeletePlugin!")

	metadata, err := meta.Accessor(input.Item)
	if err != nil {
		return err
	}

	p.log.Infof("Deleting resource: %s", metadata.GetName())

	return nil
}
