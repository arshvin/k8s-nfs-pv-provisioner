package storage

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/config"
	"testing"

	core_v1 "k8s.io/api/core/v1"
	storage_v1 "k8s.io/api/storage/v1"
)

var _appConfig *config.AppConfig

const _storageClassName = "storageClass1"

func initAppConfig() {
	_appConfig = config.GetInstance()
	_appConfig.StorageAssetRoot = "/some/path"

	scParams := map[string]string{
		"defaultOwnerAssetUid": "1000",
		"defaultOwnerAssetGid": "1000",
		"assetRoot":            "/some/path"}

	var retainPolicy = core_v1.PersistentVolumeReclaimRetain

	sc1 := new(storage_v1.StorageClass)
	sc1.Name = _storageClassName
	sc1.Provisioner = "some-vendor/some-provisioner1"
	sc1.ReclaimPolicy = &retainPolicy
	sc1.Parameters = scParams

	_appConfig.ParseStorageClass(sc1)
}

func init() {
	initAppConfig()
}

func getPvcForTests(annotations map[string]string, storageClassName string) *core_v1.PersistentVolumeClaim {
	pvc := new(core_v1.PersistentVolumeClaim)
	pvc.Name = "test-pvc"
	pvc.Spec.StorageClassName = &storageClassName
	pvc.Annotations = annotations
	return pvc
}

func checkTestResults(t *testing.T, description string, expected, actual interface{}) {
	if expected != actual {
		t.Errorf("Description: '%v', Expected value: %v but actual: %v", description, expected, actual)
	}
}

func Test_chooseBaseNameOfAsset(t *testing.T) {
	checkTestResults(t, "", "some-namespace-some-pvc-vol", ChooseBaseNameOfAsset("some-namespace", "some-pvc"))
	checkTestResults(t, "", "some-namespace-some-pvc-vol", ChooseBaseNameOfAsset("some", "namespace", "some", "pvc"))
	checkTestResults(t, "", "some-namespace-some-pvc-claim-vol", ChooseBaseNameOfAsset("some-namespace", "some-pvc-claim"))
}

func Test_castToInt(t *testing.T) {
	expected := 1000
	actual, _ := castToInt("1000")
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}

	if _, err := castToInt("FF"); err == nil {
		t.Errorf("The ERR should be not nil")
	}

	if _, err := castToInt("10O"); err == nil {
		t.Error("The ERR should be not nil")
	}
}

func Test_checkMatchTrueStr(t *testing.T) {
	checkTestResults(t, "checkMatchTrueStr(true)", true, checkMatchTrueStr("true"))
	checkTestResults(t, "checkMatchTrueStr(TRUE)", true, checkMatchTrueStr("TRUE"))
	checkTestResults(t, "checkMatchTrueStr(yes)", true, checkMatchTrueStr("yes"))
	checkTestResults(t, "checkMatchTrueStr(YES)", true, checkMatchTrueStr("YES"))

	checkTestResults(t, "checkMatchTrueStr(nope)", false, checkMatchTrueStr("nope"))
	checkTestResults(t, "checkMatchTrueStr(false)", false, checkMatchTrueStr("false"))
	checkTestResults(t, "checkMatchTrueStr(true|true)", false, checkMatchTrueStr("true|true"))
	checkTestResults(t, "checkMatchTrueStr(empty string)", false, checkMatchTrueStr(""))
}

func Test_checkMatchDNSorIPV4(t *testing.T) {
	checkTestResults(t, "10.0.0.1", true, checkMatchDNSorIPV4("10.0.0.1"))
	checkTestResults(t, "domain-1.domain-2.domain-3", true, checkMatchDNSorIPV4("domain-1.domain-2.domain-3"))

	checkTestResults(t, "domain-1.domain-2.domain_3", false, checkMatchDNSorIPV4("domain-1.domain-2.domain_3"))
	checkTestResults(t, "/domain-1.domain-2.domain", false, checkMatchDNSorIPV4("/domain-1.domain-2.domain"))
}

func Test_chooseAsset_Owner(t *testing.T) {
	annotations := map[string]string{}
	pvc1 := getPvcForTests(annotations, _storageClassName)
	uid, gid := ChooseAssetOwner(pvc1)
	checkTestResults(t, "UID was gotten from storage class params", 1000, uid)
	checkTestResults(t, "GID was gotten from storage class params", 1000, gid)

	annotations[config.AnnotationOwnerNewAssetUID] = "2000"
	annotations[config.AnnotationOwnerNewAssetGID] = "2000"
	pvc2 := getPvcForTests(annotations, _storageClassName)
	uid, gid = ChooseAssetOwner(pvc2)
	checkTestResults(t, fmt.Sprintf("UID was gotten from PVC annotation: %v", config.AnnotationOwnerNewAssetUID), 2000, uid)
	checkTestResults(t, fmt.Sprintf("GID was gotten from PVC annotation: %v", config.AnnotationOwnerNewAssetGID), 2000, gid)

	delete(annotations, config.AnnotationOwnerNewAssetUID)
	pvc3 := getPvcForTests(annotations, _storageClassName)
	uid, gid = ChooseAssetOwner(pvc3)
	checkTestResults(t, "UID was gotten from storage class params", 1000, uid)
	checkTestResults(t, fmt.Sprintf("GID was gotten from PVC annotation: %v", config.AnnotationOwnerNewAssetGID), 2000, gid)

	annotations[config.AnnotationOwnerNewAssetUID1] = "4000"
	annotations[config.AnnotationOwnerNewAssetGID1] = "4000"
	pvc4 := getPvcForTests(annotations, _storageClassName)
	uid, gid = ChooseAssetOwner(pvc4)
	checkTestResults(t, fmt.Sprintf("UID was gotten from PVC annotation: %v", config.AnnotationOwnerNewAssetUID1), 4000, uid)
	checkTestResults(t, fmt.Sprintf("GID was gotten from PVC annotation: %v", config.AnnotationOwnerNewAssetGID), 2000, gid)
}
