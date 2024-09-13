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

package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero-plugin-example/internal/plugin"
	"github.com/vmware-tanzu/velero/pkg/client"
	"github.com/vmware-tanzu/velero/pkg/plugin/framework"
	plugincommon "github.com/vmware-tanzu/velero/pkg/plugin/framework/common"
)

func main() {
	config, err := client.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error reading config file: %v\n", err)
	}
	f := client.NewFactory("plugins", config)
	framework.NewServer().
		RegisterObjectStore("example.io/object-store-plugin", newObjectStorePlugin).
		RegisterVolumeSnapshotter("example.io/volume-snapshotter-plugin", newNoOpVolumeSnapshotterPlugin).
		RegisterRestoreItemAction("example.io/restore-plugin", newRestorePlugin).
		RegisterRestoreItemActionV2("example.io/restore-pluginv2", newRestorePluginV2).
		RegisterBackupItemAction("example.io/backup-plugin", newBackupPlugin).
		RegisterBackupItemActionV2("example.io/backup-pluginv2", newBackupPluginV2).
		RegisterItemBlockAction("example.io/item-block-action-plugin", newItemBlockActionPlugin(f)).
		Serve()
}

func newBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewBackupPlugin(logger), nil
}

func newBackupPluginV2(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewBackupPluginV2(logger), nil
}

func newObjectStorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewFileObjectStore(logger), nil
}

func newRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewRestorePlugin(logger), nil
}

func newRestorePluginV2(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewRestorePluginV2(logger), nil
}

func newNoOpVolumeSnapshotterPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewNoOpVolumeSnapshotter(logger), nil
}

func newItemBlockActionPlugin(f client.Factory) plugincommon.HandlerInitializer {
	return plugin.NewItemBlockActionPlugin(f)
}
