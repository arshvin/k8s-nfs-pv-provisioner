package pv

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/checker"
	"k8s-pv-provisioner/cmd/provisioner/config"
	"k8s-pv-provisioner/cmd/provisioner/storage"
	"path"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

var appConfig = config.GetInstance()

/*Handler is the business logic method of the Controller for removal of PersistentVolumes.
Once the handler returns an error the current key will be put to the queue to be processed later. If the method returns nil the key
will be withdrawn from the queue because there is no need to do anything with it */
func Handler(indexer cache.Indexer, key string) error {
	obj, exists, err := indexer.GetByKey(key)

	if err != nil {
		klog.Errorf("Could not fetch key: %v", key)
		return err
	}

	if !exists {
		klog.Warningf("PersistentVolume does not exists anymore: %v", key)
		return nil
	}

	pv := obj.(*v1.PersistentVolume)
	checkList := checker.NewPvChecker(pv)
	checkList.PerformChecks()

	if !checkList.IsAllOK() {
		if checkList.IsReleased() && checkList.HasProperClassName() && checkList.HasProperReclaimPolicy() {
			return fmt.Errorf("Not all checks of persistentVolume have been passed for removal: %v", pv.Name)
		}
		//It's not our candidate at all. Forget about it
		return nil
	}

	storageAssetPath := path.Join(appConfig.StorageAssetRoot, pv.Spec.StorageClassName, pv.Name)
	if err := storage.DeleteStorageAsset(storageAssetPath); err != nil {
		klog.Errorf("PersistentVolume: %v deleting storage asset failed: %v", pv.Name, err)
		return err
	}

	if err := appConfig.Clientset.CoreV1().PersistentVolumes().Delete(pv.Name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	klog.V(1).Infof("PersistentVolume successfully deleted: %v", pv.Name)

	return nil
}
