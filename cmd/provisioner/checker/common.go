package checker

import (
	"k8s-pv-provisioner/cmd/provisioner/config"
)

var appConfig = config.GetInstance()

const (
	notBound = iota
	properStorageClassName
	properProvisionerAnnotation
	selectorsListEmpty
	released
	properAnnotation
	properReclaimPolicy
)
