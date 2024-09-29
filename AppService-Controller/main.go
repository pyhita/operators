package main

import (
	"flag"
	"fmt"
	"operators/AppService-Controller/pkg/generated/listers/stable/v1beta1"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	clientset "operators/AppService-Controller/pkg/generated/clientset/versioned"
	"operators/AppService-Controller/pkg/generated/informers/externalversions"
	v1beta12 "operators/AppService-Controller/pkg/generated/informers/externalversions/stable/v1beta1"
	//"operators/AppService-Controller/pkg/generated/listers/stable/v1beta1"
)

// 定义一个 Controller 结构体
type Controller struct {
	queue            workqueue.RateLimitingInterface
	appServiceLister v1beta1.AppServiceLister
	appServiceSynced cache.InformerSynced
}

// New一个结构体
func NewController(queue workqueue.RateLimitingInterface, informer v1beta12.AppServiceInformer) *Controller {

	controller := &Controller{
		queue:            queue,
		appServiceLister: informer.Lister(),
		appServiceSynced: informer.Informer().HasSynced,
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				runtime.HandleError(err)
			}
			controller.queue.Add(key)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				runtime.HandleError(err)
			}
			controller.queue.Add(key)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				runtime.HandleError(err)
			}
			controller.queue.Add(key)
		},
	})

	return controller
}

// 携程执行初始化函数
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {

	defer runtime.HandleCrash()
	// 关闭queue
	defer c.queue.ShutDown()

	fmt.Printf("Start Custom Controller")

	//等待Informer刷新缓存
	if !cache.WaitForCacheSync(stopCh, c.appServiceSynced) {
		fmt.Printf("Time out waiting caches to sync")
		return
	}

	//携程处理 queue 的程序数量
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	//chan 使 run 卡在这
	<-stopCh

	fmt.Printf("Stop Custom Controller")
}

func (c *Controller) runWorker() {
	// 一直for 循环拿 queue 里面的 key，如果queue 数组没东西了，会卡住知道有数据进入 queue
	for c.ProcessItem() {
	}
}

// 获取 key
func (c *Controller) ProcessItem() bool {
	// 从 queue 中拿 key
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	// 函数执行结束删除这个 key
	defer c.queue.Done(key)

	//执行函数的具体处理功能，如果执行失败，就放进 queue 等待下一次取出 queue 执行
	if err := c.HandlerObject(key.(string)); err != nil {
		if c.queue.NumRequeues(key) < 5 {
			c.queue.Add(key)
		}
	}

	return true
}

// 这边就是通过 queue 里面的 key 获取 indexer 里面的 object
func (c *Controller) HandlerObject(key string) error {

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	appService, err := c.appServiceLister.AppServices(namespace).Get(name)

	if err != nil {
		return err
	}

	fmt.Println(appService)
	return nil

}

// 初始化 client
func initClient() (*kubernetes.Clientset, *clientset.Clientset, error) {
	var err error
	var config *rest.Config

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(可选) kubeconfig 文件的绝对路径")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig 文件的绝对路径")
	}
	flag.Parse()

	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			return nil, nil, err
		}
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	customClient, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return k8sClient, customClient, nil
}

func main() {

	_, customClient, err := initClient()
	if err != nil {
		klog.Fatal(err)
	}

	appServiceSharedInformerFactory := externalversions.NewSharedInformerFactory(customClient, 30*time.Second)

	//生成默认的 queue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	//实例化
	controller := NewController(queue, appServiceSharedInformerFactory.Stable().V1beta1().AppServices())

	stopCh := make(chan struct{})

	//启动 informer
	appServiceSharedInformerFactory.Start(stopCh)

	//执行 Run
	go controller.Run(1, stopCh)

	defer close(stopCh)
	//select 卡住函数
	select {}
}
