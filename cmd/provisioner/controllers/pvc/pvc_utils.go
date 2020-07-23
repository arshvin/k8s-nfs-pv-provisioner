package pvc

import (
	"k8s-pv-provisioner/cmd/provisioner/config"
	"k8s-pv-provisioner/cmd/provisioner/storage"
	"path"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//TODO: Refactor this approach
func provisionPVfor(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {
	currentStorageClass := appConfig.StorageClasses[*pvc.Spec.StorageClassName]

	uid, gid := storage.ChooseAssetOwner(pvc)
	storageAssetBaseName := storage.ChooseBaseNameOfAsset(pvc.Namespace, pvc.Name)

	/*appStorageAssetPath is the full path to storage asset (folder) as it is seen or reachable from container of the provisioner*/
	appStorageAssetPath := path.Join(appConfig.StorageAssetRoot, currentStorageClass.Name, storageAssetBaseName) // e.g. -> /pv-store/nfs-class1/sbx-namespace-some-app
	/*pvStorageAssetPath is the full path to storage asset (folder) as it is seen or reachable from host OS i.e. out from of the provisioner*/
	pvStorageAssetPath := path.Join(currentStorageClass.StorageAssetRoot, storageAssetBaseName) // e.g. -> /mnt/nfs/sbx-namespace-some-app

	err := storage.CreateStorageAsset(appStorageAssetPath, uid, gid)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[config.AnnotationProvisionedBy] = currentStorageClass.Provisioner
	annotations[config.AnnotationStorageClass] = currentStorageClass.Name

	var reclaimPolicy v1.PersistentVolumeReclaimPolicy
	if value, ok := pvc.ObjectMeta.Annotations[config.AnnotationReclaimPolicy]; ok {
		reclaimPolicy = v1.PersistentVolumeReclaimPolicy(value)
	} else {
		reclaimPolicy = *currentStorageClass.ReclaimPolicy
	}

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
			StorageClassName:              currentStorageClass.Name,
			AccessModes:                   pvc.Spec.AccessModes,
			Capacity:                      pvc.Spec.Resources.Requests,
			VolumeMode:                    pvc.Spec.VolumeMode,
			PersistentVolumeReclaimPolicy: reclaimPolicy,
			ClaimRef: &v1.ObjectReference{
				Kind:      pvc.Kind,
				Name:      pvc.Name,
				Namespace: pvc.Namespace,
				UID:       pvc.UID,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Type: hostPathType,
					Path: pvStorageAssetPath,
				},
			},
		},
		Status: v1.PersistentVolumeStatus{},
	}

	return pv, nil
}
