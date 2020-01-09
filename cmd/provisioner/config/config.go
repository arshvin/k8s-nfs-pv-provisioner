package config

import (
	"strconv"

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
}

var config *appConfig

/*GetInstance is the method returning the appConfig singleton object */
func GetInstance() *appConfig {
	if config == nil {
		config = new(appConfig)
	}
	return config
}

/*ParseStorageClass method parses "k8s.io/api/storage/v1.StorageClass" struc's values*/
func (conf *appConfig) ParseStorageClass(class *storage_v1.StorageClass) {
	conf.StorageClass.Name = class.Name
	conf.StorageClass.Provisioner = class.Provisioner
	conf.StorageClass.DefaultOwnerAssetUID = (getStorageClassParameters(class, "defaultOwnerAssetUid", 1)).(int)
	conf.StorageClass.DefaultOwnerAssetGID = (getStorageClassParameters(class, "defaultOwnerAssetGid", 1)).(int)
	conf.StorageClass.StorageAssetRoot = (getStorageClassParameters(class, "assetRoot", "")).(string)
}

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

/*AppConfig is the config stucture for whole app which is supposed to be filled at the start of the programm*/
type appConfig struct {
	StorageClass     storageClass
	StorageAssetRoot string
	Clientset        *kubernetes.Clientset
}
