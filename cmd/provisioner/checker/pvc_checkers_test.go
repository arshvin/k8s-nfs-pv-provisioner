package checker

import (
	"k8s-pv-provisioner/cmd/provisioner/config"
	"testing"

	core_v1 "k8s.io/api/core/v1"
	storage_v1 "k8s.io/api/storage/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _appConfig *config.AppConfig

func getPvcForTests(annotations map[string]string, selectors meta_v1.LabelSelector, storageClassName, pvName string) *core_v1.PersistentVolumeClaim {
	pvc := new(core_v1.PersistentVolumeClaim)
	pvc.Name = "test-pvc"
	pvc.Spec.StorageClassName = &storageClassName
	pvc.Annotations = annotations
	pvc.Spec.Selector = &selectors
	pvc.Spec.VolumeName = pvName
	return pvc
}

func initAppConfig() {
	_appConfig = config.GetInstance()
	_appConfig.StorageAssetRoot = "/some/path"

	scParams := map[string]string{
		"defaultOwnerAssetUid": "1000",
		"defaultOwnerAssetGid": "1000",
		"assetRoot":            "/some/path"}

	var deletePolicy core_v1.PersistentVolumeReclaimPolicy = "Delete"
	var retainPolicy core_v1.PersistentVolumeReclaimPolicy = "Retain"

	sc1 := new(storage_v1.StorageClass)
	sc1.Name = "storageClass1"
	sc1.Provisioner = "some-vendor/some-provisioner1"
	sc1.ReclaimPolicy = &deletePolicy
	sc1.Parameters = scParams

	sc2 := new(storage_v1.StorageClass)
	sc2.Name = "storageClass2"
	sc2.Provisioner = "some-vendor/some-provisioner2"
	sc2.ReclaimPolicy = &retainPolicy
	sc2.Parameters = scParams

	sc3 := new(storage_v1.StorageClass)
	sc3.Name = "storageClass3"
	sc3.Provisioner = "some-vendor/some-provisioner3"
	sc3.ReclaimPolicy = &retainPolicy
	sc3.Parameters = scParams

	_appConfig.ParseStorageClass(sc1)
	_appConfig.ParseStorageClass(sc2)
	_appConfig.ParseStorageClass(sc3)
}

func init() {
	initAppConfig()
}

func TestCheck_notBoundCheck(t *testing.T) {
	pvc1 := getPvcForTests(nil, meta_v1.LabelSelector{}, "", "")
	ch:= &PvcChecker{pvc: pvc1}

	expected := true
	actual := ch.notBound()
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}

	pvc2 := getPvcForTests(nil, meta_v1.LabelSelector{}, "", "test-pv")
	ch = &PvcChecker{pvc: pvc2}

	expected = false
	actual = ch.notBound()
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}
}

func TestCheck_properStorageClassName(t *testing.T) {
	pvc1 := getPvcForTests(nil, meta_v1.LabelSelector{}, "", "")
	ch := &PvcChecker{pvc: pvc1}

	expected := true
	actual := ch.notBound()
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}

	pvc2 := getPvcForTests(nil, meta_v1.LabelSelector{}, "", "test-pv")
	ch = &PvcChecker{pvc: pvc2}

	expected = false
	actual = ch.notBound()
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}
}
