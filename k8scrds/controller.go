package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	lochv1 "github.com/leisheyoufu/golangstudy/k8scrds/pkg/apis/loch/v1"
	clientset "github.com/leisheyoufu/golangstudy/k8scrds/pkg/client/clientset/versioned"
	studentscheme "github.com/leisheyoufu/golangstudy/k8scrds/pkg/client/clientset/versioned/scheme"
	informers "github.com/leisheyoufu/golangstudy/k8scrds/pkg/client/informers/externalversions/loch/v1"
	listers "github.com/leisheyoufu/golangstudy/k8scrds/pkg/client/listers/loch/v1"
)

const controllerAgentName = "student-controller"

const (
	SuccessSynced = "Synced"

	MessageResourceSynced = "Student synced successfully"
)

// Controller is the controller implementation for Student resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// studentclientset is a clientset for our own API group
	studentclientset clientset.Interface

	studentsLister listers.StudentLister
	studentsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

// NewController returns a new student controller
func NewController(
	kubeclientset kubernetes.Interface,
	studentclientset clientset.Interface,
	studentInformer informers.StudentInformer) *Controller {

	utilruntime.Must(studentscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:    kubeclientset,
		studentclientset: studentclientset,
		studentsLister:   studentInformer.Lister(),
		studentsSynced:   studentInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Students"),
		recorder:         recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Student resources change
	studentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueStudent,
		UpdateFunc: func(old, new interface{}) {
			oldStudent := old.(*lochv1.Student)
			newStudent := new.(*lochv1.Student)
			if oldStudent.ResourceVersion == newStudent.ResourceVersion {
				// do nothing
				return
			}
			controller.enqueueStudent(new)
		},
		DeleteFunc: controller.enqueueStudentForDelete,
	})

	return controller
}

// Start controller service
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Start controller service and sync data")

	if ok := cache.WaitForCacheSync(stopCh, c.studentsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("worker start")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("worker started")
	<-stopCh
	klog.Info("worker end")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// 取数据处理
func (c *Controller) processNextWorkItem() bool {

	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {

			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// 在syncHandler中处理业务
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// 处理
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// 从缓存中取对象
	student, err := c.studentsLister.Students(namespace).Get(name)
	if err != nil {
		// 如果Student对象被删除了，就会走到这里，所以应该在这里加入执行
		if errors.IsNotFound(err) {
			klog.Infof("Student对象被删除，请在这里执行实际的删除业务: %s/%s ...", namespace, name)

			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list student by: %s/%s", namespace, name))

		return err
	}

	klog.Infof("这里是student对象的期望状态: %#v ...", student)
	klog.Infof("实际状态是从业务层面得到的，此处应该去的实际状态，与期望状态做对比，并根据差异做出响应(新增或者删除)")

	c.recorder.Event(student, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// 数据先放入缓存，再入队列
func (c *Controller) enqueueStudent(obj interface{}) {
	var key string
	var err error
	var students []*lochv1.Student
	// check students in local cache
	students, err = c.studentsLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
	}
	klog.Infof("Enqueue strudent, current： %#v", students)
	// check key
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	// 将key放入队列
	c.workqueue.AddRateLimited(key)
}

// 删除操作
func (c *Controller) enqueueStudentForDelete(obj interface{}) {
	var key string
	var err error
	var students []*lochv1.Student
	// check students in local cache
	students, err = c.studentsLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
	}
	klog.Infof("Denqueue strudent, current： %#v", students)
	// check key
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	//再将key放入队列
	c.workqueue.AddRateLimited(key)
}
