package pvc

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s-pv-provisioner/cmd/provisioner/config"
	"k8s-pv-provisioner/cmd/provisioner/storage"
)

func provisionPVfor(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {

	uid, gid := storage.ChooseAssetOwner(pvc)
	storageAssetBaseName := storage.ChooseBaseNameOfAsset(pvc.Namespace, pvc.Name)
	classStorageAssetRoot, err := storage.CreateStorageAsset(storageAssetBaseName, uid, gid)

	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[config.AnnotationProvisionedBy] = appConfig.StorageClass.Provisioner
	annotations[config.AnnotationStorageClass] = appConfig.StorageClass.Name

	hostPathType := new(v1.HostPathType)
	*hostPathType = v1.HostPathDirectory

	pv := &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        storageAssetBaseName,
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName:              appConfig.StorageClass.Name,
			AccessModes:                   pvc.Spec.AccessModes,
			Capacity:                      pvc.Spec.Resources.Requests,
			VolumeMode:                    pvc.Spec.VolumeMode,
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete, //TODO: Should depend on reclaimPolicy of the storage class
			ClaimRef: &v1.ObjectReference{
				Kind:      pvc.Kind,
				Name:      pvc.Name,
				Namespace: pvc.Namespace,
				UID:       pvc.UID,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Type: hostPathType,
					Path: classStorageAssetRoot,
				},
			},
		},
		Status: v1.PersistentVolumeStatus{},
	}

	return pv, nil
}
