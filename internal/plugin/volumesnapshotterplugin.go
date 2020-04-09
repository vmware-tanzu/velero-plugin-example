/*
Copyright 2018, 2019 the Velero contributors.

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
	"math/rand"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Volume keeps track of volumes created by this plugin
type Volume struct {
	volType, az string
	iops        int64
}

// Snapshot keeps track of snapshots created by this plugin
type Snapshot struct {
	volID, az string
	tags      map[string]string
}

// NoOpVolumeSnapshotter is a plugin for containing state for the blockstore
type NoOpVolumeSnapshotter struct {
	config map[string]string
	logrus.FieldLogger
	volumes   map[string]Volume
	snapshots map[string]Snapshot
}

// NewNoOpVolumeSnapshotter instantiates a NoOpVolumeSnapshotter.
func NewNoOpVolumeSnapshotter(log logrus.FieldLogger) *NoOpVolumeSnapshotter {
	return &NoOpVolumeSnapshotter{FieldLogger: log}
}

var _ velero.VolumeSnapshotter = (*NoOpVolumeSnapshotter)(nil)

// Init prepares the VolumeSnapshotter for usage using the provided map of
// configuration key-value pairs. It returns an error if the VolumeSnapshotter
// cannot be initialized from the provided config. Note that after v0.10.0, this will happen multiple times.
func (p *NoOpVolumeSnapshotter) Init(config map[string]string) error {
	p.Infof("Init called", config)
	p.config = config

	// Make sure we don't overwrite data, now that we can re-initialize the plugin
	if p.volumes == nil {
		p.volumes = make(map[string]Volume)
	}
	if p.snapshots == nil {
		p.snapshots = make(map[string]Snapshot)
	}

	return nil
}

// CreateVolumeFromSnapshot creates a new volume in the specified
// availability zone, initialized from the provided snapshot,
// and with the specified type and IOPS (if using provisioned IOPS).
func (p *NoOpVolumeSnapshotter) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
	p.Infof("CreateVolumeFromSnapshot called", snapshotID, volumeType, volumeAZ, *iops)
	var volumeID string
	for {
		volumeID := snapshotID + ".vol." + strconv.FormatUint(rand.Uint64(), 10)
		if _, ok := p.volumes[volumeID]; ok {
			// Duplicate ? Retry
			continue
		}
		break
	}

	p.volumes[volumeID] = Volume{
		volType: volumeType,
		az:      volumeAZ,
		iops:    *iops,
	}
	return volumeID, nil
}

// GetVolumeInfo returns the type and IOPS (if using provisioned IOPS) for
// the specified volume in the given availability zone.
func (p *NoOpVolumeSnapshotter) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	p.Infof("GetVolumeInfo called", volumeID, volumeAZ)
	if val, ok := p.volumes[volumeID]; ok {
		iops := val.iops
		return val.volType, &iops, nil
	}
	return "", nil, errors.New("Volume " + volumeID + " not found")
}

// IsVolumeReady Check if the volume is ready.
func (p *NoOpVolumeSnapshotter) IsVolumeReady(volumeID, volumeAZ string) (ready bool, err error) {
	p.Infof("IsVolumeReady called", volumeID, volumeAZ)
	return true, nil
}

// CreateSnapshot creates a snapshot of the specified volume, and applies any provided
// set of tags to the snapshot.
func (p *NoOpVolumeSnapshotter) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	p.Infof("CreateSnapshot called", volumeID, volumeAZ, tags)
	var snapshotID string
	for {
		snapshotID = volumeID + ".snap." + strconv.FormatUint(rand.Uint64(), 10)
		p.Infof("CreateSnapshot trying to create snapshot", snapshotID)
		if _, ok := p.snapshots[snapshotID]; ok {
			// Duplicate ? Retry
			continue
		}
		break
	}

	// Remember the "original" volume, only required for the first
	// time.
	if _, exists := p.volumes[volumeID]; !exists {
		p.volumes[volumeID] = Volume{
			volType: "orignalVolumeType",
			az:      volumeAZ,
			iops:    100,
		}
	}

	// Remember the snapshot
	p.snapshots[snapshotID] = Snapshot{volID: volumeID,
		az:   volumeAZ,
		tags: tags}

	p.Infof("CreateSnapshot returning", snapshotID)
	return snapshotID, nil
}

// DeleteSnapshot deletes the specified volume snapshot.
func (p *NoOpVolumeSnapshotter) DeleteSnapshot(snapshotID string) error {
	p.Infof("DeleteSnapshot called", snapshotID)
	delete(p.snapshots, snapshotID)
	return nil
}

// GetVolumeID returns the specific identifier for the PersistentVolume.
func (p *NoOpVolumeSnapshotter) GetVolumeID(unstructuredPV runtime.Unstructured) (string, error) {
	p.Infof("GetVolumeID called", unstructuredPV)

	pv := new(v1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return "", errors.WithStack(err)
	}

	if pv.Spec.HostPath == nil {
		return "", nil
	}

	if pv.Spec.HostPath.Path == "" {
		return "", errors.New("spec.hostPath.path not found")
	}

	return pv.Spec.HostPath.Path, nil
}

// SetVolumeID sets the specific identifier for the PersistentVolume.
func (p *NoOpVolumeSnapshotter) SetVolumeID(unstructuredPV runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	p.Infof("SetVolumeID called", unstructuredPV, volumeID)

	pv := new(v1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return nil, errors.WithStack(err)
	}

	if pv.Spec.HostPath == nil {
		return nil, errors.New("spec.hostPath.path not found")
	}

	pv.Spec.HostPath.Path = volumeID

	res, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pv)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &unstructured.Unstructured{Object: res}, nil
}
