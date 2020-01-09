package main

import (
	"testing"

	// "k8s-pv-provisioner/cmd/provisioner/config"

	"k8s-pv-provisioner/cmd/provisioner/config"
	p_pvc "k8s-pv-provisioner/cmd/provisioner/controllers/pvc"
	p_pv "k8s-pv-provisioner/cmd/provisioner/controllers/pv"
	"k8s-pv-provisioner/cmd/provisioner/storage"

	// v1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storage_v1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestChooseBaseNameOfAsset(t *testing.T) {
	actual := storage.ChooseBaseNameOfAsset("some-namespace", "some-pvc")
	expected := "some-namespace-some-pvc-vol"
	if actual != expected {
		t.Errorf("Assertion fail! Expected: %v Actual: %v", expected, actual)
	}

	actual = storage.ChooseBaseNameOfAsset("some", "namespace", "some", "pvc")
	expected = "some-namespace-some-pvc-vol"
	if actual != expected {
		t.Errorf("Assertion fail! Expected: %v Actual: %v", expected, actual)
	}

	actual = storage.ChooseBaseNameOfAsset("some-namespace", "some-pvc-claim")
	expected = "some-namespace-some-pvc-claim-vol"
	if actual != expected {
		t.Errorf("Assertion fail! Expected: %v Actual: %v", expected, actual)
	}
}

func TestCastToInt(t *testing.T) {
	actual, _ := storage.CastToInt("1000")
	expected := 1000
	if actual != expected {
		t.Errorf("Assertion fail! Expected: %v Actual: %v", expected, actual)
	}

	if _, err := storage.CastToInt("FF"); err == nil {
		t.Errorf("The ERR should be not nil")
	}

	if _, err := storage.CastToInt("10O"); err == nil {
		t.Error("The ERR should be not nil")
	}
}

func TestCheckersForProvisionClaim(t *testing.T) {
	appConfig := config.GetInstance()

	classParameters := map[string]string{
		"defaultOwnerAssetUid": "1000",
		"defaultOwnerAssetGid": "1000",
		"assetRoot":            "/some/root",
	}

	appConfig.ParseStorageClass(createTestStorageClass("test-storage-class", "test-provisioner", classParameters))

	annotations := map[string]string{
		config.AnnotationStorageClass:       appConfig.StorageClass.Name,
		config.AnnotationStorageProvisioner: appConfig.StorageClass.Provisioner,
	}

	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-claim",
			Namespace:   "test-namespace",
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &appConfig.StorageClass.Name,
			VolumeName:       "",
		},
	}

	if p_pvc.IsBoundPVC(pvc) {
		t.Errorf("The PVC %v should be valid for provisioning due to it's not bound to any PV", pvc.Name)
	}

	if !(p_pvc.HasProperStorageClassName(pvc)) {
		t.Errorf("The PVC %v should be valid for provisioning due to it has right storage class", pvc.Name)
	}

	if !(p_pvc.HasProperProvisionerAnnotation(pvc)) {
		t.Errorf("The PVC %v should be valid for provisioning due to it has right provisioner annotation", pvc.Name)
	}

	if !(p_pvc.IsSelectorsListEmpty(pvc)) {
		t.Errorf("The PVC %v should be valid for provisioning due to it has no selectors", pvc.Name)
	}

	pvcFail1 := pvc.DeepCopy()
	pvcFail1.Spec.Selector = &metav1.LabelSelector{}
	if p_pvc.IsSelectorsListEmpty(pvcFail1) {
		t.Errorf("The PVC %v with selectors should not be eligible for provisioning", pvcFail1.Name)
	}
}

// This function is needed as mocking for creating testing stuff
func createTestStorageClass(name, provisioner string, parameters map[string]string) *storage_v1.StorageClass {
	return &storage_v1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner: provisioner,
		Parameters:  parameters,
	}
}

func TestCheckersForDeprovisionPersistentVolume(t *testing.T) {
	appConfig := config.GetInstance()
	annotations := make(map[string]string)
	annotations[config.AnnotationProvisionedBy] = appConfig.StorageClass.Provisioner
	annotations[config.AnnotationStorageClass] = appConfig.StorageClass.Name

	acceesModes := []v1.PersistentVolumeAccessMode{"something"}
	requestResources := v1.ResourceList{
		v1.ResourceStorage: resource.Quantity{},
	}
	volumeMode := v1.PersistentVolumeFilesystem
	hostPathType := new(v1.HostPathType)
	*hostPathType = v1.HostPathDirectory

	pv := &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "tes-volume",
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName:              appConfig.StorageClass.Name,
			AccessModes:                   acceesModes,
			Capacity:                      requestResources,
			VolumeMode:                    &volumeMode,
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete, //TODO: Should depend on reclaimPolicy of the storage class
			ClaimRef: &v1.ObjectReference{
				Kind:      "someKind",
				Name:      "someName",
				Namespace: "someNamespace",
				UID:       "psuedoUID",
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Type: hostPathType,
					Path: "/some/where/",
				},
			},
		},
		Status: v1.PersistentVolumeStatus{
			Phase: v1.VolumeReleased,
		},
	}

	if !(p_pv.IsReleasedPV(pv)){
		t.Errorf("The PV %v should be eligible to deprovsioned because it's released", pv.Name)
	}

	if !(p_pv.HasProperReclaimPolicy(pv)){
		t.Errorf("The PV %v should be eligible to deprovsioned because it's has right reclaime policy", pv.Name)
	}

	if !(p_pv.IsProperProvisionedByAnnotation(pv)){
		t.Errorf("The PV %v should be eligible to deprovsioned because it's has right anotations", pv.Name)
	}

}
