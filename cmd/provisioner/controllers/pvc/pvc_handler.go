package pvc

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/config"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

var appConfig = config.GetInstance()

/*Handler is the business logic method of the Controller to provision PersistentVolumes for corresponding PersistentVolumeClaims*/
func Handler(indexer cache.Indexer, key string) error {
	obj, exists, err := indexer.GetByKey(key)

	if err != nil {
		klog.Errorf("Could not fetch key: %v", key)
		return err
	}

	if !exists {
		klog.Warningf("PersistentVolumeClaim does not exists anymore: %v", key)
	} else {
		pvc := obj.(*v1.PersistentVolumeClaim)

		//TODO: Refactor this block
		if !IsBoundPVC(pvc) {
			if allChecksPassed(predicates, pvc) {

				klog.V(1).Infof("PersistentVolumeClaim looks like a candidate for provisioning: %v", pvc.Name)

				pv, err := provisionPVfor(pvc)
				if err != nil {
					klog.Errorf("PersistentVolume provisioning for persistentVolumeClaim: %s failed: %s", pvc.Name, err)
					return err
				}

				klog.V(1).Infof("PersistentVolume: %v will be created and bound to persistentVolumeClaim: %v", pv.Name, pvc.Name)

				if _, err = appConfig.Clientset.CoreV1().PersistentVolumes().Create(pv); err != nil {
					return err
				}
			} else {
				if HasProperStorageClassName(pvc) {
					return fmt.Errorf("Not all checks of persistentVolumeClaim have been passed to continue provisioning: %v", pvc.Name)
				}
			}
		}
	}

	return nil
}
