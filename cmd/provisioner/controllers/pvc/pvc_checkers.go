package pvc

import (
	"fmt"
	"k8s-pv-provisioner/cmd/provisioner/config"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
)

//TODO: Refactor this module at all
var predicates = []func(*v1.PersistentVolumeClaim) bool{
	HasProperStorageClassName,
	HasProperProvisionerAnnotation,
	IsSelectorsListEmpty,
}

func IsBoundPVC(pvc *v1.PersistentVolumeClaim) bool {
	if pvc.Spec.VolumeName != "" {
		klog.V(2).Infof("PersistentVolumeClaim: %v already had been bound to the volume according to Spec.VolumeName", pvc.Name)
		return true
	}

	return false
}

func HasProperStorageClassName(pvc *v1.PersistentVolumeClaim) bool {
	if _, present := appConfig.StorageClasses[*pvc.Spec.StorageClassName]; present {
		return true
	}

	storageClasses := make([]string, 0)
	for key := range appConfig.StorageClasses {
		storageClasses = append(storageClasses, key)
	}

	klog.V(2).Infof("PersistentVolumeClaim: %v should be provisioned by another storageClass rather than on of: %v", pvc.Name, strings.Join(storageClasses, ", "))
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
	sc := appConfig.StorageClasses[*pvc.Spec.StorageClassName]
	if ok && value == sc.Provisioner {
		return true
	}

	klog.V(2).Infof("PersistentVolumeClaim: %v does not have needed annotation: %v", pvc.Name, config.AnnotationStorageProvisioner)
	return false
}

func allChecksPassed(predicates []func(*v1.PersistentVolumeClaim) bool, input *v1.PersistentVolumeClaim) bool {
	var result = true
	for _, predicate := range predicates {
		result = result && predicate(input)
	}

	return result
}
