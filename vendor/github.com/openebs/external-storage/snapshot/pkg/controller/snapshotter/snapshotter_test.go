/*
Copyright 2017 The Kubernetes Authors.

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

package snapshotter

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/fake"

	crdv1 "github.com/openebs/external-storage/snapshot/pkg/apis/volumesnapshot/v1"
	snapshotfake "github.com/openebs/external-storage/snapshot/pkg/client/clientset/versioned/fake"
	"github.com/openebs/external-storage/snapshot/pkg/cloudprovider"
	"github.com/openebs/external-storage/snapshot/pkg/controller/cache"
	"github.com/openebs/external-storage/snapshot/pkg/volume"
)

// TestPlugin methods
type TestPlugin struct {
	ShouldFail            bool
	CreateCallCount       int
	DeleteCallCount       int
	RestoreCallCount      int
	DescribeCallCount     int
	FindCallCount         int
	VolumeDeleteCallCount int
}

func (tp *TestPlugin) Init(cloudprovider.Interface) {
}

func (tp *TestPlugin) SnapshotCreate(*crdv1.VolumeSnapshot, *v1.PersistentVolume, *map[string]string) (*crdv1.VolumeSnapshotDataSource, *[]crdv1.VolumeSnapshotCondition, error) {
	tp.CreateCallCount = tp.CreateCallCount + 1
	if tp.ShouldFail {
		return nil, nil, fmt.Errorf("SnapshotCreate forced failure")
	}
	return nil, nil, nil
}

func (tp *TestPlugin) SnapshotDelete(*crdv1.VolumeSnapshotDataSource, *v1.PersistentVolume) error {
	tp.DeleteCallCount = tp.DeleteCallCount + 1
	if tp.ShouldFail {
		return fmt.Errorf("SnapshotDelete forced failure")
	}
	return nil
}

func (tp *TestPlugin) SnapshotRestore(*crdv1.VolumeSnapshotData, *v1.PersistentVolumeClaim, string, map[string]string) (*v1.PersistentVolumeSource, map[string]string, error) {
	return nil, map[string]string{}, nil
}

func (tp *TestPlugin) DescribeSnapshot(snapshotData *crdv1.VolumeSnapshotData) (snapConditions *[]crdv1.VolumeSnapshotCondition, isCompleted bool, err error) {
	return nil, true, nil
}

func (tp *TestPlugin) FindSnapshot(tags *map[string]string) (*crdv1.VolumeSnapshotDataSource, *[]crdv1.VolumeSnapshotCondition, error) {
	return nil, nil, nil
}

func (tp *TestPlugin) VolumeDelete(pv *v1.PersistentVolume) error {
	return nil
}

func fakeVolumeSnapshotDataList() *crdv1.VolumeSnapshotDataList {
	return &crdv1.VolumeSnapshotDataList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: "",
			SelfLink:        "",
		},
		Items: []crdv1.VolumeSnapshotData{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "snapshotdata-test-1",
					Namespace:         "",
					CreationTimestamp: metav1.Time{},
				},
				Spec: crdv1.VolumeSnapshotDataSpec{
					VolumeSnapshotDataSource: crdv1.VolumeSnapshotDataSource{
						HostPath: &crdv1.HostPathVolumeSnapshotSource{
							Path: "/fake/file",
						},
					},
					PersistentVolumeRef: &v1.ObjectReference{
						Kind: "PersistentVolume",
						Name: "fake-pv-1",
					},
					VolumeSnapshotRef: &v1.ObjectReference{
						Kind: "VolumeSnapshot",
						Name: "fake-snapshot-1",
					},
				},
				Status: crdv1.VolumeSnapshotDataStatus{
					Conditions: []crdv1.VolumeSnapshotDataCondition{
						{
							LastTransitionTime: metav1.Time{},
							Status:             v1.ConditionTrue,
							Type:               crdv1.VolumeSnapshotDataConditionReady,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "snapshotdata-test-2",
					Namespace:         "",
					CreationTimestamp: metav1.Time{},
				},
				Spec: crdv1.VolumeSnapshotDataSpec{
					VolumeSnapshotDataSource: crdv1.VolumeSnapshotDataSource{
						HostPath: &crdv1.HostPathVolumeSnapshotSource{
							Path: "/fake/file2",
						},
					},
					PersistentVolumeRef: &v1.ObjectReference{
						Kind: "PersistentVolume",
						Name: "fake-pv-2",
					},
					VolumeSnapshotRef: &v1.ObjectReference{
						Kind: "VolumeSnapshot",
						Name: "fake-snapshot-2",
					},
				},
				Status: crdv1.VolumeSnapshotDataStatus{
					Conditions: []crdv1.VolumeSnapshotDataCondition{
						{
							LastTransitionTime: metav1.Time{},
							Status:             v1.ConditionTrue,
							Type:               crdv1.VolumeSnapshotDataConditionReady,
						},
					},
				},
			},
		},
	}
}

func fakeVolumeSnapshotList() *crdv1.VolumeSnapshotList {
	return &crdv1.VolumeSnapshotList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: "",
			SelfLink:        "",
		},
		Items: []crdv1.VolumeSnapshot{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "snapshot-test-1",
					Namespace:         "",
					CreationTimestamp: metav1.Time{},
				},
				Spec: crdv1.VolumeSnapshotSpec{
					PersistentVolumeClaimName: "fake-pvc-1",
				},
			},
		},
	}
}

func fakeNewVolumeSnapshot() *crdv1.VolumeSnapshot {
	return &crdv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "new-snapshot-test-1",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{},
		},
		Spec: crdv1.VolumeSnapshotSpec{
			PersistentVolumeClaimName: "fake-pvc-1",
		},
	}
}

func fakePV() *v1.PersistentVolume {
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fake-pv-1",
			Namespace:         "",
			CreationTimestamp: metav1.Time{},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/fake/path",
				},
			},
		},
	}
}

func fakePVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fake-pvc-1",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName: "fake-pv-1",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: "Bound",
		},
	}
}

// Tests
func Test_NewVolumeSnapshotter(t *testing.T) {
	var tp *TestPlugin = &TestPlugin{}

	clientset := fake.NewSimpleClientset()
	asw := cache.NewActualStateOfWorld()
	plugins := map[string]volume.Plugin{"hostPath": tp}
	snapshotclient := snapshotfake.NewSimpleClientset()

	vs := NewVolumeSnapshotter(snapshotclient, clientset, asw, &plugins)
	if vs == nil {
		t.Errorf("Test failed: could not create volume snapshotter")
	}
}

func Test_getSnapshotDataFromSnapshotName(t *testing.T) {
	var tp *TestPlugin = &TestPlugin{}

	clientset := fake.NewSimpleClientset()
	asw := cache.NewActualStateOfWorld()
	plugins := map[string]volume.Plugin{"hostPath": tp}
	snapshotclient := snapshotfake.NewSimpleClientset()

	vsObj := NewVolumeSnapshotter(snapshotclient, clientset, asw, &plugins)
	if vsObj == nil {
		t.Errorf("Test failed: could not create volume snapshotter")
	}

	vs := vsObj.(*volumeSnapshotter)

	snapData := vs.getSnapshotDataFromSnapshotName("fake-snapshot-1")
	if snapData == nil {
		t.Errorf("Failure: did not find VolumeSnpshotData by VolumeSnapshot name")
	}
	if snapData.ObjectMeta.Name != "snapshotdata-test-1" {
		t.Errorf("Failure: found incorrect VolumeSnpshotData for VumeSnapshot")
	}
}

func Test_takeSnapshot(t *testing.T) {
	var tp *TestPlugin = &TestPlugin{}

	clientset := fake.NewSimpleClientset()
	asw := cache.NewActualStateOfWorld()
	plugins := map[string]volume.Plugin{"hostPath": tp}
	snapshotclient := snapshotfake.NewSimpleClientset()

	vsObj := NewVolumeSnapshotter(snapshotclient, clientset, asw, &plugins)
	if vsObj == nil {
		t.Errorf("Test failed: could not create volume snapshotter")
	}
	vs := vsObj.(*volumeSnapshotter)

	pv := fakePV()
	tags := map[string]string{
		"tag1": "tag value 1",
		"tag2": "tag value 2",
	}
	snapshot := fakeNewVolumeSnapshot()
	_, _, err = vs.takeSnapshot(snapshot, pv, &tags)
	if err != nil {
		t.Errorf("Test failed, unexpected error: %v", err)
	}
	tp.ShouldFail = true
	_, _, err = vs.takeSnapshot(snapshot, pv, &tags)
	if err != nil {
		t.Errorf("Test failed, unexpected error: %v", err)
	}

	if tp.CreateCallCount != 2 {
		t.Errorf("Test failed, expected 2 CreateSnapshot calls in plugin, got %d", tp.CreateCallCount)
	}
}

func Test_deleteSnapshot(t *testing.T) {
	var tp *TestPlugin = &TestPlugin{}

	clientset := fake.NewSimpleClientset()
	asw := cache.NewActualStateOfWorld()
	plugins := map[string]volume.Plugin{"hostPath": tp}
	snapshotclient := snapshotfake.NewSimpleClientset()

	vsObj := NewVolumeSnapshotter(snapshotclient, clientset, asw, &plugins)
	if vsObj == nil {
		t.Errorf("Test failed: could not create volume snapshotter")
	}
	vs := vsObj.(*volumeSnapshotter)

	snapDataList := fakeVolumeSnapshotDataList()
	err := vs.deleteSnapshot(&snapDataList.Items[0].Spec)
	if err != nil {
		t.Errorf("Test failed, unexpected error: %v", err)
	}
	tp.ShouldFail = true
	err = vs.deleteSnapshot(&snapDataList.Items[0].Spec)
	if err == nil {
		t.Errorf("Test failed, expected error got nil")
	}

	if tp.DeleteCallCount != 2 {
		t.Errorf("Test failed, expected 2 DeleteSnapshot calls in plugin, got %d", tp.CreateCallCount)
	}
}

func Test_createSnapshotData(t *testing.T) {
	var tp *TestPlugin = &TestPlugin{}

	clientset := fake.NewSimpleClientset(fakePVC(), fakePV())
	asw := cache.NewActualStateOfWorld()
	plugins := map[string]volume.Plugin{"hostPath": tp}
	snapshotclient := snapshotfake.NewSimpleClientset()

	vsObj := NewVolumeSnapshotter(snapshotclient, clientset, asw, &plugins)
	if vsObj == nil {
		t.Errorf("Test failed: could not create volume snapshotter")
	}
	vs := vsObj.(*volumeSnapshotter)
	//func (vs *volumeSnapshotter) createVolumeSnapshotData(snapshotName string, snapshot *crdv1.VolumeSnapshot, snapshotDataSource *crdv1.VolumeSnapshotDataSource, snapStatus *[]crdv1.VolumeSnapshotCondition) (*crdv1.VolumeSnapshotData, error) {

	snapDataSource := crdv1.VolumeSnapshotDataSource{
		HostPath: &crdv1.HostPathVolumeSnapshotSource{
			Path: "/fake/file",
		},
	}
	snapConditions := []crdv1.VolumeSnapshotCondition{
		{
			LastTransitionTime: metav1.Time{},
			Status:             v1.ConditionTrue,
			Type:               crdv1.VolumeSnapshotConditionReady,
		},
	}
	retData, err := vs.createVolumeSnapshotData("default/new-snapshot-test-1", "fake-pv-1", &snapDataSource, &snapConditions)
	if err != nil {
		t.Errorf("Test failed, unexpected error: %v", err)
	}
	if retData == nil {
		t.Errorf("Test failed: faailed to create VolumeSnapshotData")
	}
}
