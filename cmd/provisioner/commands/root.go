package commands

import (
	"flag"
	"fmt"
	appConfig "k8s-pv-provisioner/cmd/provisioner/config"
	"k8s-pv-provisioner/cmd/provisioner/controllers"
	"k8s-pv-provisioner/cmd/provisioner/controllers/pv"
	"k8s-pv-provisioner/cmd/provisioner/controllers/pvc"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

var (
	rootCmd = &cobra.Command{
		Use:   "provisioner help",
		Short: "Custom persistent volume provisioner for Kubernetes",
	}

	//storageClassName is a name of the k8s storage-class for what the provisioner will work
	storageClassName string
	/*storageAssetRoot is the directory on file system under which a new storage assets will be created for provisioned PV
	or will be looked and deleted if the PV should be deprovisioned*/
	storageAssetRoot string
	//kubectlConfig is path where actual config file for kubectl is placed for the k8s
	kubectlConfig string
	//verbosityLogging is logging level for app
	verbosityLogging int
)

func init() {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "starts the watching process for provision/deprovision persistentVolume of K8s",
	}

	serveCmd.Flags().StringVar(&storageClassName, "storage-class", "", "name of storage class to watch for (requred)")
	serveCmd.Flags().StringVar(&storageAssetRoot, "storage-asset-root", "", "directory where assets will be created  (requred)")
	serveCmd.MarkFlagRequired("storage-class")
	serveCmd.MarkFlagRequired("storage-asset-root")
	serveCmd.Run = run

	rootCmd.AddCommand(serveCmd)
	rootCmd.PersistentFlags().StringVarP(&kubectlConfig, "kubectl-config", "c", "", "path to kubectl's config")
	rootCmd.PersistentFlags().IntVar(&verbosityLogging, "v", 0, "logging verbosity")
	rootCmd.PersistentPreRun = persistentPreRun
}

//Execute is the entrypoint for the RootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	//Initing the klog
	flags := flag.NewFlagSet("Logging flags", flag.ExitOnError)
	klog.InitFlags(flags)

	flags.Set("logtostderr", "true")
	flags.Set("v", strconv.Itoa(verbosityLogging))
	flags.Parse(nil)
}

func run(cmd *cobra.Command, args []string) {

	var config *rest.Config
	var err error

	if kubectlConfig == "" {
		klog.Info("Trying to use in-cluster config")
		config, err = rest.InClusterConfig()
	} else {
		klog.Info("Trying to use config specifyied as file path")
		config, err = clientcmd.BuildConfigFromFlags("", kubectlConfig)
	}

	if err != nil {
		klog.Fatal(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err.Error())
	}

	//From this point we are ready to request a data from k8s cluster
	appConfig := appConfig.GetInstance()
	appConfig.StorageAssetRoot = storageAssetRoot
	appConfig.Clientset = clientset

	storageClasses, err := clientset.StorageV1().StorageClasses().List(meta_v1.ListOptions{})
	if err != nil {
		klog.Fatal("Could not fetch list of storage classes")
	}

	for _, item := range storageClasses.Items {
		if item.Name == storageClassName {
			appConfig.ParseStorageClass(&item)
			break
		}
	}
	if appConfig.StorageClass.Name == "" {
		klog.Fatalf("Could not find specified storage class: %v", storageClassName)
	}

	//Preparing stuff for PVC controller
	pvcListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "persistentvolumeclaims", meta_v1.NamespaceAll, fields.Everything())
	pvcQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	pvcIndexer, pvcInformer := cache.NewIndexerInformer(pvcListWatcher, &v1.PersistentVolumeClaim{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.V(4).Infof("Added object: %v", obj)
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				pvcQueue.Add(key)
				klog.V(2).Infof("The new persistentVolumeClaim was added: %v", key)
			}
		},
	}, cache.Indexers{})

	//Preparing stuff for PV controller
	pvListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "persistentvolumes", meta_v1.NamespaceNone, fields.Everything())
	pvQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	pvIndexer, pvInformer := cache.NewIndexerInformer(pvListWatcher, &v1.PersistentVolume{}, 0, cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			klog.V(4).Infof("Changed object: %v", newObj)
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				pvQueue.Add(key)
				klog.V(2).Infof("The persistentVolume was changed: %v", key)
			}
		},
	}, cache.Indexers{})

	var pvcCtrl = controllers.NewController("PersistentVolumeClaim", pvcQueue, pvcIndexer, pvcInformer)
	pvcCtrl.ItemHandler = pvc.Handler

	var pvCtrl = controllers.NewController("PersistentVolume", pvQueue, pvIndexer, pvInformer)
	pvCtrl.ItemHandler = pv.Handler

	//Starting the controllers with one stop-channel
	stop := make(chan struct{})
	defer close(stop)

	go pvcCtrl.Run(stop)
	go pvCtrl.Run(stop)

	//Wait forever
	select {}
}
