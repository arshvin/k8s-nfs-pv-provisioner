package checker

import (
	"k8s-pv-provisioner/cmd/provisioner/config"
	"testing"

	core_v1 "k8s.io/api/core/v1"
	storage_v1 "k8s.io/api/storage/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _appConfig *config.AppConfig

func getPvcForTests(annotations map[string]string, selectors *meta_v1.LabelSelector, storageClassName, pvName string) *core_v1.PersistentVolumeClaim {
	pvc := new(core_v1.PersistentVolumeClaim)
	pvc.Name = "test-pvc"
	pvc.Spec.StorageClassName = &storageClassName
	pvc.Annotations = annotations
	pvc.Spec.Selector = selectors
	pvc.Spec.VolumeName = pvName
	return pvc
}

func getPvForTests(annotations map[string]string, reclaimPolicy core_v1.PersistentVolumeReclaimPolicy, storageClassName, pvName string, phase core_v1.PersistentVolumePhase) *core_v1.PersistentVolume {
	pv := new(core_v1.PersistentVolume)
	pv.Annotations = annotations
	pv.Spec.StorageClassName = storageClassName
	pv.Spec.PersistentVolumeReclaimPolicy = reclaimPolicy
	pv.Status.Phase = phase
	pv.Name = pvName
	return pv
}

func checkTestResults(t *testing.T, expected, actual bool) {
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}
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

func TestPVC_check_notBoundCheck(t *testing.T) {
	pvc1 := getPvcForTests(nil, nil, "", "")
	pvc2 := getPvcForTests(nil, nil, "", "test-pv")

	ch := NewPvcChecker(pvc1)
	checkTestResults(t, true, ch.notBound())

	ch = NewPvcChecker(pvc2)
	checkTestResults(t, false, ch.notBound())
}

func TestPVC_check_properStorageClassName(t *testing.T) {
	pvc1 := getPvcForTests(nil, nil, "storageClass0", "")
	pvc2 := getPvcForTests(nil, nil, "storageClass1", "test-pv")

	ch := NewPvcChecker(pvc1)
	checkTestResults(t, false, ch.properStorageClassName())

	ch = NewPvcChecker(pvc2)
	checkTestResults(t, true, ch.properStorageClassName())
}

func TestPVC_check_selectorsListEmpty(t *testing.T) {
	selector := &meta_v1.LabelSelector{MatchLabels: map[string]string{"key1": "value1"}}
	pvc1 := getPvcForTests(nil, selector, "", "")
	pvc2 := getPvcForTests(nil, nil, "", "")

	ch := NewPvcChecker(pvc1)
	checkTestResults(t, false, ch.selectorsListEmpty())

	ch = NewPvcChecker(pvc2)
	checkTestResults(t, true, ch.selectorsListEmpty())
}

func TestPVC_check_properProvisionerAnnotation(t *testing.T) {
	annotations := map[string]string{
		"volume.beta.kubernetes.io/storage-provisioner": "some-vendor/some-provisioner2",
	}

	pvc1 := getPvcForTests(annotations, nil, "storageClass1", "")
	pvc2 := getPvcForTests(annotations, nil, "storageClass2", "")

	ch := NewPvcChecker(pvc1)
	checkTestResults(t, false, ch.properProvisionerAnnotation())

	ch = NewPvcChecker(pvc2)
	checkTestResults(t, true, ch.properProvisionerAnnotation())

}

func TestPVC_IsAllOK(t *testing.T) {
	annotations := map[string]string{
		"volume.beta.kubernetes.io/storage-provisioner": "some-vendor/some-provisioner1",
	}
	pvc1 := getPvcForTests(annotations, nil, "storageClass1", "")
	pvc2 := getPvcForTests(annotations, nil, "storageClass1", "some-pv")

	ch := NewPvcChecker(pvc1)
	ch.PerformChecks()
	checkTestResults(t, true, ch.IsAllOK())

	ch = NewPvcChecker(pvc2)
	ch.PerformChecks()
	checkTestResults(t, false, ch.IsAllOK())
}

func TestPV_check_properReclaimPolicy(t *testing.T) {
	pv1 := getPvForTests(nil, core_v1.PersistentVolumeReclaimRetain, "", "", "")
	pv2 := getPvForTests(nil, core_v1.PersistentVolumeReclaimRecycle, "", "", "")
	pv3 := getPvForTests(nil, core_v1.PersistentVolumeReclaimDelete, "", "", "")

	ch := NewPvChecker(pv1)
	checkTestResults(t, false, ch.properReclaimPolicy())

	ch = NewPvChecker(pv2)
	checkTestResults(t, false, ch.properReclaimPolicy())

	ch = NewPvChecker(pv3)
	checkTestResults(t, true, ch.properReclaimPolicy())
}

func TestPV_check_properClassName(t *testing.T) {
	pv1 := getPvForTests(nil, "", "storageClass0", "", "")
	pv2 := getPvForTests(nil, "", "storageClass1", "", "")

	ch := NewPvChecker(pv1)
	checkTestResults(t, false, ch.properClassName())

	ch = NewPvChecker(pv2)
	checkTestResults(t, true, ch.properClassName())
}

func TestPV_check_properAnnotations(t *testing.T) {
	annotations := map[string]string{
		"pv.kubernetes.io/provisioned-by": "some-vendor/some-provisioner1",
	}
	pv1 := getPvForTests(annotations, "", "storageClass1", "", "")

	annotations = map[string]string{
		"pv.kubernetes.io/provisioned-by": "some-vendor/some-provisioner3",
	}
	pv2 := getPvForTests(annotations, "", "storageClass2", "", "")

	annotations = map[string]string{
		"pv.kubernetes.io/provisioned-by": "some-vendor/some-provisioner4",
	}
	pv3 := getPvForTests(annotations, "", "storageClass4", "", "")

	ch := NewPvChecker(pv1)
	checkTestResults(t, true, ch.properAnnotations())

	ch = NewPvChecker(pv2)
	checkTestResults(t, false, ch.properAnnotations())

	ch = NewPvChecker(pv3)
	checkTestResults(t, false, ch.properAnnotations())
}

func TestPV_check_released(t *testing.T) {
	pv1 := getPvForTests(nil, "", "", "", core_v1.VolumeBound)
	pv2 := getPvForTests(nil, "", "", "", core_v1.VolumeReleased)

	ch := NewPvChecker(pv1)
	checkTestResults(t, false, ch.released())

	ch = NewPvChecker(pv2)
	checkTestResults(t, true, ch.released())
}

func TestPV_check_IsAllOk(t *testing.T) {
	annotations := map[string]string{
		"pv.kubernetes.io/provisioned-by": "some-vendor/some-provisioner1",
	}
	deletePolicy := core_v1.PersistentVolumeReclaimDelete
	retainPolicy := core_v1.PersistentVolumeReclaimRetain
	releasedPhase := core_v1.VolumeReleased

	pv1 := getPvForTests(annotations, deletePolicy, "storageClass1", "", releasedPhase)
	ch := NewPvChecker(pv1)
	ch.PerformChecks()
	checkTestResults(t, true, ch.IsAllOK())

	pv2 := getPvForTests(annotations, retainPolicy, "storageClass1", "", releasedPhase)
	ch = NewPvChecker(pv2)
	ch.PerformChecks()
	checkTestResults(t, false, ch.IsAllOK())
}
