package storage

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"k8s-pv-provisioner/cmd/provisioner/config"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

var appConfig = config.GetInstance()

//ChooseBaseNameOfAsset is func which returns the base name of new storagge asset depending on input args
func ChooseBaseNameOfAsset(args ...string) string {
	result := make([]string, len(args))
	for index, item := range args {
		item = strings.Trim(item, "-")
		result[index] = item
	}

	result = append(result, "vol")
	return strings.Join(result, "-") //-> arg1-arg2...-argN-vol
}

//ChooseAssetOwner is func which decides of what uid gid attributes the new asset should have
func ChooseAssetOwner(pvc *v1.PersistentVolumeClaim) (int, int) {
	var uid, gid int
	var err error

	tmp, ok := pvc.Annotations[config.AnnotationOwnerNewAssetUID]
	if ok {
		uid, err = CastToInt(tmp)
		if err != nil {
			uid = appConfig.StorageClass.DefaultOwnerAssetUID
			klog.Warningf("PersistentVolumeClaim: %v annotation: %v could not parse value: %v: %v", pvc.Name, config.AnnotationOwnerNewAssetUID, tmp, err)
		}
	} else {
		uid = appConfig.StorageClass.DefaultOwnerAssetUID
	}

	tmp, ok = pvc.Annotations[config.AnnotationOwnerNewAssetGID]
	if ok {
		gid, err = CastToInt(tmp)
		if err != nil {
			gid = appConfig.StorageClass.DefaultOwnerAssetGID
			klog.Warningf("PersistentVolumeClaim: %v annotation: %v could not parse value: %v: %v", pvc.Name, config.AnnotationOwnerNewAssetGID, tmp, err)
		}
	} else {
		gid = appConfig.StorageClass.DefaultOwnerAssetGID
	}

	return uid, gid
}

func CastToInt(value string) (int, error) {
	id, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

//CreateStorageAsset is func which creates the storage asset
func CreateStorageAsset(baseName string, uid, gid int) (string, error) {

	appStorageAssetRoot := path.Join(appConfig.StorageAssetRoot, baseName)
	classStorageAssetRoot := path.Join(appConfig.StorageClass.StorageAssetRoot, baseName)

	if _, err := os.Stat(appStorageAssetRoot); err == nil {
		return "", fmt.Errorf("Storage asset: %v already exists", appStorageAssetRoot)
	}

	if err := os.MkdirAll(appStorageAssetRoot, 0755); os.IsPermission(err) {
		return "", err
	}

	klog.Infof("Storage asset: %v was successfully created", appStorageAssetRoot)

	if err := os.Chown(appStorageAssetRoot, uid, gid); err != nil {

		return "", err
	}

	klog.Infof("Storage asset: %v ownership was set as %v:%v", appStorageAssetRoot, uid, gid)

	return classStorageAssetRoot, nil
}

//DeleteStorageAsset is func deleting the storage asset
func DeleteStorageAsset(pv *v1.PersistentVolume) error {
	classStorageAsset := pv.Spec.PersistentVolumeSource.HostPath.Path
	appStorageAsset := path.Join(appConfig.StorageAssetRoot, path.Base(classStorageAsset))

	err := os.RemoveAll(appStorageAsset)
	if err != nil {
		return err
	}

	klog.Infof("Storage asset: %v was successfully deleted", appStorageAsset)

	return nil
}
