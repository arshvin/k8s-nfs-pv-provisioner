package pv

import (
	"k8s-pv-provisioner/cmd/provisioner/config"
	"k8s-pv-provisioner/cmd/provisioner/storage"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

var appConfig = config.GetInstance()

/*Handler is the business logic method of the Controller for deprovisioning of released PersistentVolumes*/
func Handler(indexer cache.Indexer, key string) error {
	obj, exists, err := indexer.GetByKey(key)

	if err != nil {
		klog.Errorf("Could not fetch key: %v", key)
		return err
	}

	if !exists {
		klog.Warningf("PersistentVolume does not exists anymore: %v", key)
	} else {
		pv := obj.(*v1.PersistentVolume)

		if IsReleasedPV(pv) {
			if allChecksPassed(predicates, pv) {

				if err := storage.DeleteStorageAsset(pv); err != nil {
					klog.Errorf("PersistentVolume: %v deleting storage asset failed: %v", pv.Name, err)
					return err
				}

				if err := appConfig.Clientset.CoreV1().PersistentVolumes().Delete(pv.Name, &metav1.DeleteOptions{}); err != nil {
					return nil
				}
			}
		}
	}

	return nil
}
