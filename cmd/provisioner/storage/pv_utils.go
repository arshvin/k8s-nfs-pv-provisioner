package storage

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/config"
	"path"
	"strings"

	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type pvArguments struct {
	name          string
	pvc           *core_v1.PersistentVolumeClaim
	reclaimPolicy core_v1.PersistentVolumeReclaimPolicy
	annotations   map[string]string
	assetPath     string
}

/*PreparePV is function which creates storage asset(folder) and returns prepared PV structure to be created in cluster. Depending on presence colon sign in
StorageAssetRoot field of currentStorageClass NFS or HostPath type of PV will be returned*/
func PreparePV(pvc *core_v1.PersistentVolumeClaim) (*core_v1.PersistentVolume, error) {
	currentStorageClass := appConfig.StorageClasses[*pvc.Spec.StorageClassName]

	uid, gid := ChooseAssetOwner(pvc)
	storageAssetBaseName := ChooseBaseNameOfAsset(pvc.Namespace, pvc.Name)

	/*appStorageAssetPath is the full path to storage asset (folder) as it is seen or reachable from container of the provisioner*/
	appStorageAssetPath := path.Join(appConfig.StorageAssetRoot, currentStorageClass.Name, storageAssetBaseName) // e.g. -> /pv-store/nfs-class1/sbx-namespace-some-app
	/*pvStorageAssetPath is the full path to storage asset (folder) as it is seen or reachable from host OS i.e. out from of the provisioner*/
	pvStorageAssetPath := path.Join(currentStorageClass.StorageAssetRoot, storageAssetBaseName) // e.g. -> /mnt/nfs/sbx-namespace-some-app

	var reuseExistingAsset bool
	if _, present := pvc.Annotations[config.AnnotationUseExistingAsset]; present {
		reuseExistingAsset = true
	}
	err := CreateStorageAsset(appStorageAssetPath, uid, gid, reuseExistingAsset)
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

	pv := fillPV(pvArgs)

	//This can happened after panic recovery
	if pv == nil {
		if !reuseExistingAsset {
			/*If we are here and reuseExistingAsset == false than we can be sure that storage asset was created by us a few lines earlier.
			On next iteration when the provioner will try to provision this PV it will face with issue that the storage asset already exists.
			Therefore because of this issue we must delete created storage asset in current iteration when the panic occured*/
			DeleteStorageAsset(appStorageAssetPath)
		}

		return pv, fmt.Errorf("Could not prepare new PV")
	}

	return pv, nil
}

func getHostPathPersistentVolumeSource(assetPath string) core_v1.PersistentVolumeSource {
	hostPathType := new(core_v1.HostPathType)
	*hostPathType = core_v1.HostPathDirectory

	return core_v1.PersistentVolumeSource{
		HostPath: &core_v1.HostPathVolumeSource{
			Type: hostPathType,
			Path: assetPath,
		},
	}
}
func getNfsPersistentVolumeSource(assetPath string) core_v1.PersistentVolumeSource {
	nfsAssetPath := strings.Split(assetPath, ":")
	if len(nfsAssetPath) > 2 {
		panic(fmt.Sprintf("A storage class assetRoot must contain only 1 colon sign if NFS-like path usage is assumed. Got value: %v", assetPath))
	}
	//TODO: Here should be some checks whether the nfsAssetPath[0] meet hostname reqirements. Implement them
	return core_v1.PersistentVolumeSource{
		NFS: &core_v1.NFSVolumeSource{
			Server:   nfsAssetPath[0],
			Path:     nfsAssetPath[1],
			ReadOnly: false,
		},
	}
}

func fillPV(args *pvArguments) *core_v1.PersistentVolume {
	defer func() {
		if err := recover(); err != nil {
			klog.Error(err)
		}
	}()

	var persistentVolumeSource core_v1.PersistentVolumeSource
	if strings.Contains(args.assetPath, ":") {
		persistentVolumeSource = getNfsPersistentVolumeSource(args.assetPath)
	} else {
		persistentVolumeSource = getHostPathPersistentVolumeSource(args.assetPath)
	}

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
			PersistentVolumeSource: persistentVolumeSource,
		},
		Status: core_v1.PersistentVolumeStatus{},
	}
}
