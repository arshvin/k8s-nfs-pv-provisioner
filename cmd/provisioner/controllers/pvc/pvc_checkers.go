package pvc

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/config"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
)

var predicates = []func(*v1.PersistentVolumeClaim) bool{
	HasProperStorageClassName,
	HasProperProvisionerAnnotation,
	IsSelectorsListEmpty,
}

func IsBoundPVC(pvc *v1.PersistentVolumeClaim) bool {
	if pvc.Spec.VolumeName != "" {
		klog.V(2).Infof("PersistentVolumeClaim: %v has been bound to a volume according to Spec.VolumeName", pvc.Name)
		return true
	}

	return false
}

func HasProperStorageClassName(pvc *v1.PersistentVolumeClaim) bool {
	if *pvc.Spec.StorageClassName == appConfig.StorageClass.Name {
		return true
	}

	klog.V(2).Infof("PersistentVolumeClaim: %v should be provisioned by another storageClass than: %v", pvc.Name, appConfig.StorageClass.Name)
	return false
}

func IsSelectorsListEmpty(pvc *v1.PersistentVolumeClaim) bool {
	if pvc.Spec.Selector != nil {
		runtime.HandleError(fmt.Errorf("PersistentVolumeClaim: %v should not have selectors in order to be provisioned", pvc.Name))
		return false
	}

	return true
}

func HasProperProvisionerAnnotation(pvc *v1.PersistentVolumeClaim) bool {
	value, ok := pvc.Annotations[config.AnnotationStorageProvisioner]
	if ok && value == appConfig.StorageClass.Provisioner {
		return true
	}

	klog.V(2).Infof("PersistentVolumeClaim: %v does not have annotation specifying onto the provisioner: %v", pvc.Name, config.AnnotationStorageProvisioner)
	return false
}

func allChecksPassed(predicates []func(*v1.PersistentVolumeClaim) bool, input *v1.PersistentVolumeClaim) bool {
	var result = true
	for _, predicate := range predicates {
		result = result && predicate(input)
	}

	return result
}
