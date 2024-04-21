package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listcorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

/**
  应用中有多处相互独立的业务逻辑都需要监控同一种资源对象，用户会编写多个informer来处理
  这会导致应用中发起对K8sAPIserver同一种资源的多次listandwatch调用，
  并且每个informer中都有一份单独的本地缓存，着呢宫颈癌了内存的开销
  使用SharedInformer后，客户端对同一种资源对象只会有一个对APIserver的ListAndWatc调用，
  多个informer共用一份缓存，减少对apiserver的请求，增加了系统的性能
*/

type Controller struct {
	lister   listcorev1.PodLister
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

func NewController(queue workqueue.RateLimitingInterface, lister listcorev1.PodLister, informer cache.Controller) *Controller {
	return &Controller{
		lister:   lister,
		queue:    queue,
		informer: informer,
	}
}

func (c *Controller) Run(workers int, stopCh chan struct{}) {
	defer utilruntime.HandleCrash()

	defer c.queue.ShuttingDown()
	klog.Info("starting pod controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timeout waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)

	}

	<-stopCh
	klog.Info("stopping pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncToStdout(key.(string))
	c.handleErr(err, key)
	return true
}

func (c Controller) syncToStdout(key string) error {

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invaild resource key: %s", key))
		return nil
	}

	pod, err := c.lister.Pods(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("pod %s not in workqueue", key))
			return nil
		}
		return err
	}

	fmt.Printf("sync/add/update for pod %s\n", pod.GetName())
	return nil
}

func (c *Controller) handleErr(err error, key interface{}) {

	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("error syncing pod %v:%v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	utilruntime.HandleError(err)
	klog.Infof("dropping pod %q out of the queue: %v", key, err)

}

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", ".kube/kubeconfig")
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	podInformer := informerFactory.Core().V1().Pods()

	//create the queue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	//register the event handler
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	controller := NewController(queue, podInformer.Lister(), podInformer.Informer())

	stopCh := make(chan struct{})
	defer close(stopCh)

	//启动informer监听
	informerFactory.Start(stopCh)

	go controller.Run(1, stopCh)

	select {}
}
