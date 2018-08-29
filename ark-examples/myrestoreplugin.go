/*
Copyright 2018 the Heptio Ark contributors.

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

package main

import (
	"github.com/heptio/ark/pkg/apis/ark/v1"
	"github.com/heptio/ark/pkg/restore"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// MyRestorePlugin is a restore item action plugin for Heptio Ark
type MyRestorePlugin struct {
	log logrus.FieldLogger
}

// AppliesTo returns a restore.ResourceSelector that applies to everything
func (p *MyRestorePlugin) AppliesTo() (restore.ResourceSelector, error) {
	return restore.ResourceSelector{}, nil
}

func (p *MyRestorePlugin) Execute(item runtime.Unstructured, restore *v1.Restore) (runtime.Unstructured, error, error) {
	p.log.Info("Hello from MyRestorePlugin!")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["ark.heptio.com/my-restore-plugin"] = "1"

	metadata.SetAnnotations(annotations)

	return item, nil, nil
}
