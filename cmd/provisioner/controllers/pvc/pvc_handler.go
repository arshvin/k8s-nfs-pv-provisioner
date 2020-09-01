package pvc

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/checker"
	"k8s-pv-provisioner/cmd/provisioner/config"
	"k8s-pv-provisioner/cmd/provisioner/storage"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

var appConfig = config.GetInstance()

/*Handler is the business logic method of the Controller to provision PersistentVolumes for corresponding PersistentVolumeClaims.
Once the handler returns an error the current key will be put to the queue to be processed later. If the method returns nil the key
will be withdrawn from the queue because there is no need to do anything with it */
func Handler(indexer cache.Indexer, key string) error {
	obj, exists, err := indexer.GetByKey(key)

	if err != nil {
		klog.Errorf("Could not fetch key: %v", key)
		return err
	}

	if !exists {
		klog.Warningf("PersistentVolumeClaim does not exists anymore: %v", key)
		return nil
	}

	pvc := obj.(*core_v1.PersistentVolumeClaim)
	checkList := checker.NewPvcChecker(pvc)
	checkList.PerformChecks()

	if !checkList.IsAllOK() {
		if checkList.IsNotBound() && checkList.HasProperStorageClassName() {
			return fmt.Errorf("Not all checks of persistentVolumeClaim have been passed to continue provisioning: %v", pvc.Name)
		}
		//It's not our canditate at all. Forget about it
		return nil
	}

	klog.V(1).Infof("PersistentVolumeClaim looks like a candidate for provisioning: %v", pvc.Name)

	pv, err := storage.PreparePV(pvc)
	if err != nil {
		klog.Errorf("PersistentVolume provisioning for persistentVolumeClaim: %s failed: %s", pvc.Name, err)
		return err
	}

	if _, err = appConfig.Clientset.CoreV1().PersistentVolumes().Create(pv); err != nil {
		return err
	}

	klog.V(1).Infof("PersistentVolume: %v successfully created and bound to persistentVolumeClaim: %v", pv.Name, pvc.Name)

	return nil
}
