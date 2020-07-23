package config

import (
	"strconv"

	core_v1 "k8s.io/api/core/v1"
	storage_v1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

/*storageClass is set of the needed items for work fetched from "k8s.io/api/storage/v1" and few custom ones*/
type storageClass struct {
	//Name is the name of the storage class that will be specifeid in PersistentVolumeSpec
	Name string
	//DefaultOwnerAssetUID is default uid for new created assests if there are no annotations in PVCs overriding it
	DefaultOwnerAssetUID int
	//DefaultOwnerAssetGID is default gid for new created assests if there are no annotations in PVCs overriding it
	DefaultOwnerAssetGID int
	//StorageAssetRoot is full path where new assets will be creted and used by provisioned PVs
	StorageAssetRoot string
	//Provisioner is name of the provisioner that will be specified in annotations for provisioned PVs
	Provisioner string
	//ReclaimPolicy is value of *v1.PersistentVolumeReclaimPolicy which should have one of Recycle, Delete, Retain
	ReclaimPolicy *core_v1.PersistentVolumeReclaimPolicy
}

var config *AppConfig

/*GetInstance is the method returning the appConfig singleton object */
func GetInstance() *AppConfig {
	if config == nil {
		config = new(AppConfig)
		config.StorageClasses = make(StorageClassesMap)
	}
	return config
}

/*ParseStorageClass is a method for parsing "k8s.io/api/storage/v1.StorageClass" struc to fill appConfig.ParseStorageClass map*/
func (conf *AppConfig) ParseStorageClass(class *storage_v1.StorageClass) {
	sc := new(storageClass)
	sc.Name = class.Name
	sc.Provisioner = class.Provisioner
	sc.ReclaimPolicy = class.ReclaimPolicy
	sc.DefaultOwnerAssetUID = (getStorageClassParameters(class, "defaultOwnerAssetUid", 1)).(int)
	sc.DefaultOwnerAssetGID = (getStorageClassParameters(class, "defaultOwnerAssetGid", 1)).(int)
	sc.StorageAssetRoot = (getStorageClassParameters(class, "assetRoot", "")).(string)

	conf.StorageClasses[sc.Name] = *sc
}

//TODO: Refactor this function
func getStorageClassParameters(class *storage_v1.StorageClass, paramName string, castTo interface{}) interface{} {
	tmpStr, ok := class.Parameters[paramName]
	if !ok {
		klog.Fatalf("Parameter '%s' in storage class '%s' must be defined", paramName, class.Name)
	}

	switch castTo.(type) {
	case int:
		tmpInt, err := strconv.ParseInt(tmpStr, 10, 32)
		if err != nil {
			klog.Fatalf("Could not cast to UInt value of the parameter '%s': %v", paramName, tmpStr)
		}
		return int(tmpInt)
	case string:
		return tmpStr
	default:
		return nil
	}
}

/*StorageClassesMap is the map of storage classes that the provisioner will serve*/
type StorageClassesMap map[string]storageClass

/*AppConfig is the config stucture for whole app which is supposed to be filled at the start of the programm*/
type AppConfig struct {
	StorageClasses   StorageClassesMap
	StorageAssetRoot string
	Clientset        *kubernetes.Clientset
}
