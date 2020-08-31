package pvc

import (
	"k8s-pv-provisioner/cmd/provisioner/config"
	"k8s-pv-provisioner/cmd/provisioner/storage"
	"path"

	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	
)

type pvArguments struct {
	name          string
	pvc           *core_v1.PersistentVolumeClaim
	reclaimPolicy core_v1.PersistentVolumeReclaimPolicy
	annotations   map[string]string
	assetPath     string
}

func preparePv(pvc *core_v1.PersistentVolumeClaim) (*core_v1.PersistentVolume, error) {
	currentStorageClass := appConfig.StorageClasses[*pvc.Spec.StorageClassName]

	uid, gid := storage.ChooseAssetOwner(pvc)
	storageAssetBaseName := storage.ChooseBaseNameOfAsset(pvc.Namespace, pvc.Name)

	/*appStorageAssetPath is the full path to storage asset (folder) as it is seen or reachable from container of the provisioner*/
	appStorageAssetPath := path.Join(appConfig.StorageAssetRoot, currentStorageClass.Name, storageAssetBaseName) // e.g. -> /pv-store/nfs-class1/sbx-namespace-some-app
	/*pvStorageAssetPath is the full path to storage asset (folder) as it is seen or reachable from host OS i.e. out from of the provisioner*/
	pvStorageAssetPath := path.Join(currentStorageClass.StorageAssetRoot, storageAssetBaseName) // e.g. -> /mnt/nfs/sbx-namespace-some-app

	var reuseExistingAsset bool
	if _, present := pvc.Annotations[config.AnnotationUseExistingAsset]; present {
		reuseExistingAsset = true
	}
	err := storage.CreateStorageAsset(appStorageAssetPath, uid, gid, reuseExistingAsset)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[config.AnnotationProvisionedBy] = currentStorageClass.Provisioner
	annotations[config.AnnotationStorageClass] = currentStorageClass.Name

	var reclaimPolicy core_v1.PersistentVolumeReclaimPolicy
	if value, ok := pvc.Annotations[config.AnnotationReclaimPolicy]; ok {
		reclaimPolicy = core_v1.PersistentVolumeReclaimPolicy(value)
	} else {
		reclaimPolicy = *currentStorageClass.ReclaimPolicy
	}

	pvArgs := new(pvArguments)
	pvArgs.name = storageAssetBaseName
	pvArgs.assetPath = pvStorageAssetPath
	pvArgs.annotations = annotations
	pvArgs.reclaimPolicy = reclaimPolicy
	pvArgs.pvc = pvc

	pv := getHostPathPV(pvArgs)
	return pv, nil
}

func getHostPathPV(args *pvArguments) *core_v1.PersistentVolume {
	hostPathType := new(core_v1.HostPathType)
	*hostPathType = core_v1.HostPathDirectory

	return &core_v1.PersistentVolume{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        args.name,
			Annotations: args.annotations,
		},
		Spec: core_v1.PersistentVolumeSpec{
			StorageClassName:              *args.pvc.Spec.StorageClassName,
			AccessModes:                   args.pvc.Spec.AccessModes,
			Capacity:                      args.pvc.Spec.Resources.Requests,
			VolumeMode:                    args.pvc.Spec.VolumeMode,
			PersistentVolumeReclaimPolicy: args.reclaimPolicy,
			ClaimRef: &core_v1.ObjectReference{
				Kind:      args.pvc.Kind,
				Name:      args.pvc.Name,
				Namespace: args.pvc.Namespace,
				UID:       args.pvc.UID,
			},
			PersistentVolumeSource: core_v1.PersistentVolumeSource{
				HostPath: &core_v1.HostPathVolumeSource{
					Type: hostPathType,
					Path: args.assetPath,
				},
			},
		},
		Status: core_v1.PersistentVolumeStatus{},
	}
}
