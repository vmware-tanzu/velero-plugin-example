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
	veleroplugin "github.com/heptio/velero/pkg/plugin/framework"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer().
		RegisterBackupItemAction("example/backup-plugin", newBackupPlugin).
		RegisterObjectStore("example/object-store-plugin", newObjectStorePlugin).
		RegisterRestoreItemAction("example/restore-plugin", newRestorePlugin).
		RegisterVolumeSnapshotter("example/volume-snapshotter-plugin", newNoOpVolumeSnapshotterPlugin).
		Serve()
}

func newBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &BackupPlugin{log: logger}, nil
}

func newObjectStorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &FileObjectStore{log: logger}, nil
}

func newRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &RestorePlugin{log: logger}, nil
}

func newNoOpVolumeSnapshotterPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &NoOpVolumeSnapshotter{FieldLogger: logger}, nil
}
