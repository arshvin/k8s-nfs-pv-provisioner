package controllers

import (
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

var (
	resourceType = map[string]runtime.Object{
		"persistentvolumeclaims": &core_v1.PersistentVolumeClaim{},
		"persistentvolumes":      &core_v1.PersistentVolume{},
	}
)

//PrepareStuff is the function that returns all stuff that is needed to launch controller
func PrepareStuff(clientset *kubernetes.Clientset, resource string) (workqueue.RateLimitingInterface, cache.Indexer, cache.Controller) {
	listWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), resource, meta_v1.NamespaceAll, fields.Everything())
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	var eventHandler cache.ResourceEventHandlerFuncs
	switch resource {
	case "persistentvolumeclaims":
		eventHandler = cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				klog.V(3).Infof("Added object: %v", obj)
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
					klog.V(2).Infof("The new persistentVolumeClaim was added: %v", key)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				klog.V(3).Infof("Changed object: %v", newObj)
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err == nil {
					queue.Add(key)
					klog.V(2).Infof("The persistentVolumeClaim was changed: %v", key)
				}
			},
		}
	case "persistentvolumes":
		eventHandler = cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				klog.V(3).Infof("Changed object: %v", newObj)
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err == nil {
					queue.Add(key)
					klog.V(2).Infof("The persistentVolume was changed: %v", key)
				}
			},
		}
	}

	indexer, informer := cache.NewIndexerInformer(listWatcher, resourceType[resource], 0, eventHandler, cache.Indexers{})

	return queue, indexer, informer
}
