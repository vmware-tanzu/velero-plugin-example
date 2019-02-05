package main

import (
	"errors"
	"math/rand"
	"strconv"

	"github.com/heptio/velero/pkg/util/collections"
	"github.com/heptio/velero/pkg/cloudprovider"
	"github.com/sirupsen/logrus"
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

// Plugin for containing state for the blockstore plugin
type NoOpBlockStore struct {
	config map[string]string
	logrus.FieldLogger
	volumes   map[string]Volume
	snapshots map[string]Snapshot
}

var _ cloudprovider.BlockStore = (*NoOpBlockStore)(nil)

// Init the plugin. Note that after v0.10.0, this will happen multiple times.
func (p *NoOpBlockStore) Init(config map[string]string) error {
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

// CreateVolumeFromSnapshot Create a volume from given snapshot.
func (p *NoOpBlockStore) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
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

// GetVolumeInfo Get information about the volume.
func (p *NoOpBlockStore) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	p.Infof("GetVolumeInfo called", volumeID, volumeAZ)
	if val, ok := p.volumes[volumeID]; ok {
		iops := val.iops
		return val.volType, &iops, nil
	}
	return "", nil, errors.New("Volume " + volumeID + " not found")
}

// IsVolumeReady Check if the volume is ready.
func (p *NoOpBlockStore) IsVolumeReady(volumeID, volumeAZ string) (ready bool, err error) {
	p.Infof("IsVolumeReady called", volumeID, volumeAZ)
	return true, nil
}

// CreateSnapshot Create a snapshot.
func (p *NoOpBlockStore) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
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

// DeleteSnapshot Delete a snapshot.
func (p *NoOpBlockStore) DeleteSnapshot(snapshotID string) error {
	p.Infof("DeleteSnapshot called", snapshotID)
	delete(p.snapshots, snapshotID)
	return nil
}

// GetVolumeID Get the volume ID from the spec.
func (p *NoOpBlockStore) GetVolumeID(pv runtime.Unstructured) (string, error) {
	p.Infof("GetVolumeID called", pv)
	if !collections.Exists(pv.UnstructuredContent(), "spec.hostPath.path") {
		return "", errors.New("Example plugin failed to get volume ID. ")
	}

	// Seed the volume info so that GetVolumeInfo doesn't fail later.
	volumeID, _ := collections.GetString(pv.UnstructuredContent(), "spec.hostPath.path")
	if _, exists := p.volumes[volumeID]; !exists {
		p.Infof("L134")
		p.volumes[volumeID] = Volume{
			volType: "orignalVolumeType",
			iops:    100,
		}
	}

	return collections.GetString(pv.UnstructuredContent(), "spec.hostPath.path")
}

// SetVolumeID Set the volume ID in the spec.
func (p *NoOpBlockStore) SetVolumeID(pv runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	p.Infof("SetVolumeID called", pv, volumeID)
	metadataMap, err := collections.GetMap(pv.UnstructuredContent(), "spec.hostPath.path")
	if err != nil {
		return nil, err
	}

	metadataMap["volumeID"] = volumeID
	return pv, nil
}
