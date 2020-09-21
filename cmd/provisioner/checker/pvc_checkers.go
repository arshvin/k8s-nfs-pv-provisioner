package checker

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/config"
	"strings"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
)

/*PvcChecker is a gatekeeper through which a PVC should pass to reach to PV provision*/
type PvcChecker struct {
	AbstractChecker
	pvc *core_v1.PersistentVolumeClaim
}

func (ch PvcChecker) notBound() bool {
	if ch.pvc.Spec.VolumeName == "" {
		return true
	}

	klog.V(2).Infof("PersistentVolumeClaim: %v already had been bound to the volume according to Spec.VolumeName", ch.pvc.Name)
	return false
}

func (ch PvcChecker) properStorageClassName() bool {
	if _, ok := appConfig.StorageClasses[*ch.pvc.Spec.StorageClassName]; ok {
		return true
	}

	classNames := make([]string, len(appConfig.StorageClasses))
	index := 0
	for key := range appConfig.StorageClasses {
		classNames[index] = key
		index++
	}

	klog.V(2).Infof("PersistentVolumeClaim: %v should be provisioned by another storageClass rather than: %v", ch.pvc.Name, strings.Join(classNames, ", "))
	return false
}

func (ch PvcChecker) selectorsListEmpty() bool {
	if ch.pvc.Spec.Selector != nil {
		runtime.HandleError(fmt.Errorf("PersistentVolumeClaim: %v must not have selectors in order to be provisioned", ch.pvc.Name))
		return false
	}
	return true
}

func (ch PvcChecker) properProvisionerAnnotation() bool {
	sc := appConfig.StorageClasses[*ch.pvc.Spec.StorageClassName]
	value, ok := ch.pvc.Annotations[config.AnnotationStorageProvisioner]
	if ok && value == sc.Provisioner {
		return true
	}

	klog.V(2).Infof("PersistentVolumeClaim: %v does not have needed annotation: %v", ch.pvc.Name, config.AnnotationStorageProvisioner)
	return false
}

//IsNotBound is method returning true if the PVC is not bound yet or false otherwise
func (ch PvcChecker) IsNotBound() bool {
	return ch.AbstractChecker.Results[notBound]
}

//HasProperStorageClassName is method returning whether PVC has proper storage class to be served by the provisioner
func (ch PvcChecker) HasProperStorageClassName() bool {
	return ch.AbstractChecker.Results[properStorageClassName]
}

//NewPvcChecker is the factory function for creation PvcChecker
func NewPvcChecker(pvc *core_v1.PersistentVolumeClaim) *PvcChecker {
	ch := new(PvcChecker)
	ch.pvc = pvc
	ch.AbstractChecker.Checker = ch
	return ch
}

func (ch PvcChecker) checkList() map[int]func() bool {
	return map[int]func() bool{
		properStorageClassName:      ch.properStorageClassName,
		properProvisionerAnnotation: ch.properProvisionerAnnotation,
		notBound:                    ch.notBound,
		selectorsListEmpty:          ch.selectorsListEmpty,
	}
}
