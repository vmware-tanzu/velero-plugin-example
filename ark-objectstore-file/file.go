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
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type FileObjectStore struct {
	log logrus.FieldLogger
}

func (f *FileObjectStore) Init(config map[string]string) error {
	f.log.Infof("FileObjectStore.Init called")
	return os.MkdirAll(getRoot(), 0755)
}

func (f *FileObjectStore) PutObject(bucket string, key string, body io.Reader) error {
	path := filepath.Join(getRoot(), bucket, key)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"path":   path,
	})
	log.Infof("PutObject")

	dir := filepath.Dir(path)
	log.Infof("Creating dir %s", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	log.Infof("Creating file")
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Infof("Writing to file")
	_, err = io.Copy(file, body)

	log.Infof("Done")
	return err
}

func (f *FileObjectStore) GetObject(bucket string, key string) (io.ReadCloser, error) {
	path := filepath.Join(getRoot(), bucket, key)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"path":   path,
	})
	log.Infof("GetObject")

	return os.Open(path)
}

func (f *FileObjectStore) ListCommonPrefixes(bucket string, delimiter string) ([]string, error) {
	path := filepath.Join(getRoot(), bucket, delimiter)

	log := f.log.WithFields(logrus.Fields{
		"bucket":    bucket,
		"delimiter": delimiter,
		"path":      path,
	})
	log.Infof("ListCommonPrefixes")

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, info := range infos {
		if info.IsDir() {
			dirs = append(dirs, info.Name())
		}
	}

	return dirs, nil
}

func (f *FileObjectStore) ListObjects(bucket string, prefix string) ([]string, error) {
	path := filepath.Join(getRoot(), bucket, prefix)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"prefix": prefix,
		"path":   path,
	})
	log.Infof("ListObjects")

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var objects []string
	for _, info := range infos {
		objects = append(objects, info.Name())
	}

	return objects, nil
}

func (f *FileObjectStore) DeleteObject(bucket string, key string) error {
	path := filepath.Join(getRoot(), bucket, key)

	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"path":   path,
	})
	log.Infof("DeleteObject")

	return os.Remove(path)
}

func (f *FileObjectStore) CreateSignedURL(bucket string, key string, ttl time.Duration) (string, error) {
	return "", errors.New("CreateSignedURL is not supported for this plugin")
}

const defaultRoot = "/tmp/backups"

func getRoot() string {
	root := os.Getenv("ARK_FILE_OBJECT_STORE_ROOT")
	if root != "" {
		return root
	}

	return defaultRoot
}
