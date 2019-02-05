/*
Copyright 2017 the Heptio Ark contributors.

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
	veleroplugin "github.com/heptio/velero/pkg/plugin"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer(veleroplugin.NewLogger()).
		RegisterBackupItemAction("backup-plugin", newBackupPlugin).
		RegisterObjectStore("file", newFileObjectStore).
		RegisterRestoreItemAction("restore-plugin", newMyRestorePlugin).
		RegisterBlockStore("example-blockstore", newNoOpBlockStore).
		Serve()
}

func newBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &BackupPlugin{log: logger}, nil
}

func newFileObjectStore(logger logrus.FieldLogger) (interface{}, error) {
	return &FileObjectStore{log: logger}, nil
}

func newMyRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return &MyRestorePlugin{log: logger}, nil
}

func newNoOpBlockStore(logger logrus.FieldLogger) (interface{}, error) {
	return &NoOpBlockStore{FieldLogger: logger}, nil
}
