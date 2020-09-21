package checker

import (
	"k8s-pv-provisioner/cmd/provisioner/config"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)


/*PvChecker is a gatekeeper though which a PV should pass to be deleted*/
type PvChecker struct {
	AbstractChecker
	pv      *core_v1.PersistentVolume
}

func (ch PvChecker) released() bool {
	if ch.pv.Status.Phase == core_v1.VolumeReleased {
		return true
	}

	klog.V(2).Infof("PersistentVolume: %v is not released yet", ch.pv.Name)
	return false
}

func (ch PvChecker) properAnnotations() bool {
	storageClassName := ch.pv.Spec.StorageClassName
	currentStorageClass := appConfig.StorageClasses[storageClassName]

	value, ok := ch.pv.Annotations[config.AnnotationProvisionedBy]
	if ok && value == currentStorageClass.Provisioner {
		return true
	}

	klog.V(2).Infof("PersistentVolume: %v does not have right annotation: %v", ch.pv.Name, currentStorageClass.Provisioner)
	return false
}

func (ch PvChecker) properReclaimPolicy() bool {
	if ch.pv.Spec.PersistentVolumeReclaimPolicy == core_v1.PersistentVolumeReclaimDelete {
		return true
	}

	klog.V(2).Infof("PersistentVolume: %v does not have right reclaimPolicy", ch.pv.Name)
	return false

}

func (ch PvChecker) properClassName() bool {
	storageClassName := ch.pv.Spec.StorageClassName
	if _, present := appConfig.StorageClasses[storageClassName]; present {
		return true
	}

	klog.V(2).Infof("StorageClass: %v of PersistentVolume: %v is not served by current provioner", storageClassName, ch.pv.Name)
	return false
}

//IsReleased is method returning the result of whether the PV released or not to be delete by the provisioner
func (ch PvChecker) IsReleased() bool {
	return ch.Results[released]
}

//HasProperClassName is method returning whether PV has proper storage class to be delete by the provisioner
func (ch PvChecker) HasProperClassName() bool {
	return ch.Results[properStorageClassName]
}

//HasProperReclaimPolicy is method returning whether PV has proper reclaim to be delete by the provisioner
func (ch PvChecker) HasProperReclaimPolicy() bool {
	return ch.Results[properReclaimPolicy]
}

//NewPvChecker is the factory function for creation PvChecker
func NewPvChecker(pv *core_v1.PersistentVolume) *PvChecker {
	ch := new(PvChecker)
	ch.pv = pv
	ch.AbstractChecker.Checker = ch
	return ch
}

func (ch PvChecker) checkList() map[int]func() bool {
	return map[int]func() bool{
		properStorageClassName: ch.properClassName,
		properAnnotation:       ch.properAnnotations,
		released:               ch.released,
		properReclaimPolicy:    ch.properReclaimPolicy,
	}
}
