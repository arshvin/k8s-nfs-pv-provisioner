package checker

import (
	"k8s-pv-provisioner/cmd/provisioner/config"
)

var appConfig = config.GetInstance()

const (
	notBound                    = 1 << iota
	properStorageClassName      = 1 << iota
	properProvisionerAnnotation = 1 << iota
	selectorsListEmpty          = 1 << iota
	released                    = 1 << iota
	properAnnotation            = 1 << iota
	properReclaimPolicy         = 1 << iota
)
