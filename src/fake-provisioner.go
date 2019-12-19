package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	storage_v1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

type AppConfig struct {
	storageClass      string
	innerAssetRoot    string
	kubectlConfigPath string
}

type Provisioner struct {
	storageClass       *storage_v1.StorageClass
	name               string
	innerAssetRootPath string
	outerAssetRootPath string
	ownerNewAssetUID   int
	ownerNewAssetGID   int
}

var (
	buffer      = os.Stdout
	logger      = log.New(buffer, "", log.Ldate|log.Ltime|log.Lshortfile)
	provisioner Provisioner
)

const (
	sleepMilliseconds            = 1000
	outerAssetRootParameter      = "outerAssetRoot"
	ownerNewAssetUIDParameter    = "ownerNewAssetUid"
	ownerNewAssetGIDParameter    = "ownerNewAssetGid"
	annotationProvisionedBy      = "pv.kubernetes.io/provisioned-by"
	annotationStorageClass       = "volume.beta.kubernetes.io/storage-class"
	annotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)

func (provisioner *Provisioner) init(client *kubernetes.Clientset, appConfig AppConfig) {
	storageClasses, err := client.StorageV1().StorageClasses().List(metav1.ListOptions{})

	if err != nil {
		logger.Panic("Could not get the list of storge classes")
	}

	for _, storageClassName := range storageClasses.Items {
		if storageClassName.Name == appConfig.storageClass {
			provisioner.storageClass = &storageClassName
			break
		}
	}

	if provisioner.storageClass == nil {
		logger.Panicf("Specifyed storage class '%s' not found in the cluster", appConfig.storageClass)
	}

	provisioner.name = provisioner.storageClass.Provisioner

	assetRoot, ok := provisioner.storageClass.Parameters[outerAssetRootParameter]
	if !ok {
		logger.Panicf("The parameter '%s' in storage class '%s' must be defined", outerAssetRootParameter, appConfig.storageClass)
	}

	tmpStrVar, ok := provisioner.storageClass.Parameters[ownerNewAssetUIDParameter]
	if !ok {
		logger.Panicf("The parameter '%s' in storage class '%s' must be defined", ownerNewAssetUIDParameter, appConfig.storageClass)
	}

	ownerNewAssetUID, err := strconv.ParseInt(tmpStrVar, 10, 32)
	if err != nil {
		logger.Panicf("Could not cast to UInt value of the parameter '%s'", ownerNewAssetUIDParameter)
	}

	tmpStrVar, ok = provisioner.storageClass.Parameters[ownerNewAssetGIDParameter]
	if !ok {
		logger.Panicf("The parameter '%s' of storage class '%s' must be defined", ownerNewAssetGIDParameter, appConfig.storageClass)
	}

	ownerNewAssetGID, err := strconv.ParseInt(tmpStrVar, 10, 32)
	if err != nil {
		logger.Panicf("Could not cast to UInt value of the parameter '%s'", ownerNewAssetGIDParameter)
	}

	provisioner.outerAssetRootPath = assetRoot
	provisioner.innerAssetRootPath = appConfig.innerAssetRoot
	provisioner.ownerNewAssetUID = int(ownerNewAssetUID)
	provisioner.ownerNewAssetGID = int(ownerNewAssetGID)
}

func (provisioner Provisioner) createPV(client *kubernetes.Clientset, pvcChannel <-chan v1.PersistentVolumeClaim) {
	createStorageAsset := func(storageAsset string) (string, string, error) {

		storageAssetInner := path.Join(provisioner.innerAssetRootPath, storageAsset)
		storageAssetOuter := path.Join(provisioner.outerAssetRootPath, storageAsset)

		if _, err := os.Stat(storageAssetInner); err == nil {
			return "", "", fmt.Errorf("The asset '%s'(container) or '%s'(OS) already exists", storageAssetInner, storageAssetOuter)
		}

		if err := os.MkdirAll(storageAssetInner, 0777); os.IsPermission(err) {
			return "", "", fmt.Errorf("Permission denied to create the asset '%s'(container) or '%s'(OS)", storageAssetInner, storageAssetOuter)
		}

		if err := os.Chown(storageAssetInner, provisioner.ownerNewAssetUID, provisioner.ownerNewAssetGID); err != nil {
			return "", "", fmt.Errorf("Ownership setup for the asset '%s'(container) or '%s'(OS) was failed", storageAssetInner, storageAssetOuter)
		}

		logger.Printf("Storage asset '%s' was successfully created", storageAssetOuter)
		return storageAssetInner, storageAssetOuter, nil
	}

	for pvc := range pvcChannel {
		var stringBuilder strings.Builder
		//If claim has in its own name the word "claim" we will delete it
		fmt.Fprintf(&stringBuilder, "%s-%s-vol", pvc.Namespace, strings.ReplaceAll(pvc.Name, "claim", ""))
		storageAsset := stringBuilder.String()

		_, storageAssetOuter, err := createStorageAsset(storageAsset)
		if err != nil {
			logger.Printf("PV provision for claim '%s' failed: %s", pvc.Name, err)
			continue //Skip further work for the claim
		}

		pv := new(v1.PersistentVolume)
		pvMeta := &pv.ObjectMeta
		pvMeta.SetName(storageAsset)

		annotations := make(map[string]string)
		annotations[annotationProvisionedBy] = provisioner.name
		annotations[annotationStorageClass] = provisioner.storageClass.Name
		pvMeta.SetAnnotations(annotations)

		pvcRef := new(v1.ObjectReference)
		pvcRef.Kind = pvc.Kind
		pvcRef.Name = pvc.Name
		pvcRef.Namespace = pvc.Namespace
		pvcRef.UID = pvc.UID

		pvSpec := &pv.Spec
		pvSpec.StorageClassName = provisioner.storageClass.Name
		pvSpec.AccessModes = pvc.Spec.AccessModes
		pvSpec.Capacity = pvc.Spec.Resources.Requests
		pvSpec.VolumeMode = pvc.Spec.VolumeMode
		pvSpec.PersistentVolumeReclaimPolicy = v1.PersistentVolumeReclaimDelete

		pvHostPathVolumeSource := new(v1.HostPathVolumeSource)
		pvHostPathVolumeSource.Path = storageAssetOuter
		pvHostPathVolumeSource.Type = new(v1.HostPathType)
		*pvHostPathVolumeSource.Type = v1.HostPathDirectory

		pvPersistentVolumeSource := &pv.Spec.PersistentVolumeSource
		pvPersistentVolumeSource.HostPath = pvHostPathVolumeSource

		pvSpec.ClaimRef = pvcRef

		_, err = client.CoreV1().PersistentVolumes().Create(pv)
		if err != nil {
			logger.Printf("Creating PV '%s' and binding to PVC '%s' failed: %s", pv.Name, pvc.Name, err.Error())
		}

		logger.Printf("PV '%s' was created and bound to PVC '%s'", pv.Name, pvc.Name)
	}
}

func (provisioner Provisioner) deletePV(client *kubernetes.Clientset, pvChannel <-chan v1.PersistentVolume) {
	deleteStorageAsset := func(pv v1.PersistentVolume) error {
		storageAssetOuter := pv.Spec.PersistentVolumeSource.HostPath.Path
		storageAssetInner := path.Join(provisioner.innerAssetRootPath, path.Base(storageAssetOuter))

		err := os.RemoveAll(storageAssetInner)
		if err != nil {
			return err
		}
		logger.Printf("Storage asset '%s' was successfully deleted", storageAssetOuter)

		return nil
	}

	for pv := range pvChannel {
		err := deleteStorageAsset(pv)
		if err != nil {
			logger.Printf("Could not delete storage asset of PV '%s': %s ", pv.Name, err.Error())
			continue
		}

		err = client.CoreV1().PersistentVolumes().Delete(pv.Name, &metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Could not delete PV '%s': %s", pv.Name, err.Error())
		}
		logger.Printf("PV '%s' was successfully deleted", pv.Name)

	}
}

func (provisioner Provisioner) lookingPvcToProvision(client *kubernetes.Clientset, pvcChannel chan<- v1.PersistentVolumeClaim) {
	isEligibleClaim := func(pvc *v1.PersistentVolumeClaim) bool {
		if pvc.Annotations[annotationStorageProvisioner] != provisioner.name {
			return false
		}

		if pvc.Spec.VolumeName != "" {
			return false
		}

		if pvc.Annotations[annotationStorageClass] != provisioner.storageClass.Name && *pvc.Spec.StorageClassName != provisioner.storageClass.Name {
			return false
		}

		if pvc.Spec.Selector != nil {
			runtime.HandleError(fmt.Errorf("Could not parse '%s' claim' selectors. The claim will not be provisioned", pvc.Name))
			return false
		}

		return true
	}

	namespaces, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		logger.Print(err.Error())
		return
	}

	for _, namespace := range namespaces.Items {
		name := namespace.Name
		pvc, err := client.CoreV1().PersistentVolumeClaims(name).List(metav1.ListOptions{})
		if err != nil {
			logger.Print(err.Error())
			return
		}

		for _, claim := range pvc.Items {
			if isEligibleClaim(&claim) {
				pvcChannel <- claim
				logger.Printf("Claim '%s' was chosen as eligible and will be provisioned", claim.Name)
			}
		}
	}
}

func (provisioner Provisioner) lookingPvToDelete(client *kubernetes.Clientset, pvChannel chan<- v1.PersistentVolume) {
	isEligiblePv := func(pv *v1.PersistentVolume) bool {
		if pv.Status.Phase != v1.VolumeReleased {
			return false
		}

		if pv.Annotations[annotationProvisionedBy] != provisioner.name {
			return false
		}

		if pv.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimDelete {
			return false
		}

		return true
	}

	volumes, err := client.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		logger.Print(err.Error())
		return
	}

	for _, pv := range volumes.Items {
		if isEligiblePv(&pv) {
			pvChannel <- pv
		}
	}
}

func main() {

	var appConfig AppConfig

	flag.StringVar(&appConfig.innerAssetRoot, "inner-asset-root", "", "Absolute path in container where network share directory mapped into")
	flag.StringVar(&appConfig.storageClass, "storage-class", "", "'Storage class name' of claims to looking for")
	flag.StringVar(&appConfig.kubectlConfigPath, "kubectl-config", "", "Absolute path to kubectl's config")
	flag.Parse()

	if appConfig.storageClass == "" {
		logger.Fatal("Argument storage-class must be specified")
	}

	if appConfig.innerAssetRoot == "" {
		logger.Fatal("Argument inner-asset-root must be specified")
	}

	var config *rest.Config
	var err error

	if appConfig.kubectlConfigPath == "" {
		logger.Print("Trying to use in-cluster config")
		config, err = rest.InClusterConfig()
	} else {
		logger.Print("Trying to use config specifyied as file path")
		config, err = clientcmd.BuildConfigFromFlags("", appConfig.kubectlConfigPath)
	}

	if err != nil {
		logger.Fatal(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Fatal(err.Error())
	}

	provisioner := new(Provisioner)
	provisioner.init(clientset, appConfig)

	pvcChannel := make(chan v1.PersistentVolumeClaim)
	pvChannel := make(chan v1.PersistentVolume)

	//Does not make a sence since all channels will be closed when the app will exit, but let it be
	defer func() {
		close(pvcChannel)
		close(pvChannel)
	}()

	go provisioner.createPV(clientset, pvcChannel)
	go provisioner.deletePV(clientset, pvChannel)

	for {
		provisioner.lookingPvcToProvision(clientset, pvcChannel)
		provisioner.lookingPvToDelete(clientset, pvChannel)

		time.Sleep(sleepMilliseconds * time.Millisecond)
	}
}
