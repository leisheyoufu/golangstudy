///*
//
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
//package controllers
//
//import (
//	"context"
//	"fmt"
//	"reflect"
//	"time"
//
//	"github.com/go-logr/logr"
//	corev1 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/errors"
//	"k8s.io/apimachinery/pkg/labels"
//	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
//
//	//"k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	ctrl "sigs.k8s.io/controller-runtime"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//
//	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
//
//	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
//	"sigs.k8s.io/controller-runtime/pkg/reconcile"
//
//	dbv1 "github.com/example-inc/mysql-operator/api/v1"
//	"k8s.io/apimachinery/pkg/util/intstr"
//)
//
//// MysqlReconciler reconciles a Mysql object
//type MysqlReconciler struct {
//	client.Client
//	Log    logr.Logger
//	Scheme *runtime.Scheme
//}
//
//// +kubebuilder:rbac:groups=db.example.com,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
//// +kubebuilder:rbac:groups=db.example.com,resources=mysqls/status,verbs=get;update;patch
//
//func (r *MysqlReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
//	_ = context.Background()
//	_ = r.Log.WithValues("mysql", req.NamespacedName)
//	var err error
//	var ret ctrl.Result
//	var pod *corev1.Pod
//	var svc *corev1.Service
//	//your logic here
//	instance := &dbv1.Mysql{}
//	err = r.Get(context.TODO(), req.NamespacedName, instance)
//	if err != nil {
//		r.Log.Info("Mysql resource name", "ResourceName", req.NamespacedName, "Error", err)
//		if errors.IsNotFound(err) {
//			return reconcile.Result{}, nil
//		}
//		return ctrl.Result{}, err
//	}
//	ret, err = r.Finalize(instance)
//	if err != nil {
//		return reconcile.Result{}, err
//	}
//	pod, ret, err = r.reconcilePod(instance)
//	if err != nil {
//		return reconcile.Result{}, err
//	}
//	if ret.Requeue == true {
//		return ret, nil
//	}
//	svc, ret, err = r.reconcileService(instance)
//	if err != nil {
//		return reconcile.Result{}, err
//	}
//	if ret.Requeue == true {
//		return ret, nil
//	}
//	node, err := r.getNode()
//	if err != nil {
//		return reconcile.Result{}, err
//	}
//	status := dbv1.MysqlStatus{
//		PodName:  pod.Name,
//		Endpoint: fmt.Sprintf("%s:%d", node.Status.Addresses[0].Address, svc.Spec.Ports[0].NodePort),
//	}
//	if !reflect.DeepEqual(instance.Status, status) {
//		instance.Status = status
//		err = r.Status().Update(context.TODO(), instance)
//		if err != nil {
//			r.Log.Error(err, "Failed to update mysql status")
//			return reconcile.Result{}, err
//		}
//	}
//	return reconcile.Result{}, nil
//}
//
//func (r *MysqlReconciler) Finalize(instance *dbv1.Mysql) (ctrl.Result, error) {
//	finalizerName := "storage.finalizers.tutorial.kubebuilder.io"
//	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
//		// 如果为0 ，则资源未被删除，我们需要检测是否存在 finalizer，如果不存在，则添加，并更新到资源对象中
//		if !containsString(instance.ObjectMeta.Finalizers, finalizerName) {
//			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
//			if err := r.Update(context.Background(), instance); err != nil {
//				return ctrl.Result{}, err
//			}
//		}
//	} else {
//		// 如果不为 0 ，则对象处于删除中
//		if containsString(instance.ObjectMeta.Finalizers, finalizerName) {
//			// 如果存在 finalizer 且与上述声明的 finalizer 匹配，那么执行对应 hook 逻辑
//			if err := r.deleteExternalResources(instance); err != nil {
//				// 如果删除失败，则直接返回对应 err，controller 会自动执行重试逻辑
//				return ctrl.Result{}, err
//			}
//
//			// 如果对应 hook 执行成功，那么清空 finalizers， k8s 删除对应资源
//			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
//			if err := r.Update(context.Background(), instance); err != nil {
//				return ctrl.Result{}, err
//			}
//		}
//	}
//	return ctrl.Result{}, nil
//}
//
//func (r *MysqlReconciler) SetupWithManager(mgr ctrl.Manager) error {
//	return ctrl.NewControllerManagedBy(mgr).
//		For(&dbv1.Mysql{}).
//		Owns(&corev1.Pod{}).
//		Owns(&corev1.Service{}).
//		Complete(r)
//}
//
//func (r *MysqlReconciler) getNode() (*corev1.Node, error) {
//	var err error
//	nodes := &corev1.NodeList{}
//	err = r.List(context.TODO(), nodes)
//	if err != nil {
//		return nil, err
//	}
//	if len(nodes.Items) == 0 {
//		return nil, errors.NewInternalError(fmt.Errorf("No available nodes found"))
//	}
//	return &nodes.Items[0], nil
//}
//
//func (r *MysqlReconciler) reconcileService(instance *dbv1.Mysql) (*corev1.Service, ctrl.Result, error) {
//	var err error
//	svcs := &corev1.ServiceList{}
//	lbs := map[string]string{
//		"app":     instance.Name,
//		"version": instance.Spec.Version,
//	}
//	labelSelector := labels.SelectorFromSet(lbs)
//	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
//	if err := r.List(context.TODO(), svcs, listOps); err != nil {
//		return nil, reconcile.Result{}, err
//	}
//	if len(svcs.Items) > 1 {
//		err = r.Delete(context.TODO(), &svcs.Items[1])
//		if err != nil {
//			return nil, reconcile.Result{}, err
//		}
//		return nil, reconcile.Result{Requeue: true}, nil
//	}
//	if len(svcs.Items) == 0 {
//		svc := &corev1.Service{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:         instance.Name + "-svc",
//				GenerateName: instance.Name + "-svc",
//				Namespace:    instance.Namespace,
//				Labels:       lbs,
//			},
//			Spec: corev1.ServiceSpec{
//				Type:     corev1.ServiceTypeNodePort,
//				Selector: lbs,
//				Ports: []corev1.ServicePort{
//					{
//						Name:       instance.Name,
//						Protocol:   corev1.ProtocolTCP,
//						NodePort:   instance.Spec.Port,
//						TargetPort: intstr.FromInt(3306),
//						Port:       3306,
//					},
//				},
//			},
//		}
//		if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
//			return nil, reconcile.Result{}, err
//		}
//		err = r.Create(context.TODO(), svc)
//		if err != nil {
//			return nil, reconcile.Result{}, err
//		}
//		return nil, reconcile.Result{Requeue: true}, nil
//	}
//	return &svcs.Items[0], reconcile.Result{}, nil
//}
//
//func (r *MysqlReconciler) reconcilePod(instance *dbv1.Mysql) (*corev1.Pod, ctrl.Result, error) {
//	var err error
//	pods := &corev1.PodList{}
//	lbs := map[string]string{
//		"app":     instance.Name,
//		"version": instance.Spec.Version,
//	}
//	labelSelector := labels.SelectorFromSet(lbs)
//	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
//	if err = r.List(context.TODO(), pods, listOps); err != nil {
//		return nil, reconcile.Result{}, err
//	}
//	var available []corev1.Pod
//	for _, pod := range pods.Items {
//		if pod.ObjectMeta.DeletionTimestamp != nil {
//			continue
//		}
//		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
//			available = append(available, pod)
//		}
//	}
//	numAvailable := int32(len(available))
//	if numAvailable > 1 {
//		err = r.Delete(context.TODO(), &pods.Items[1])
//		if err != nil {
//			return nil, reconcile.Result{}, err
//		}
//	}
//	// Update the status if necessary
//	if len(pods.Items) == 0 {
//		r.Log.Info("Scaling up pods", "Currently available", numAvailable, "Required replicas", 1)
//		pod := newPodForCR(instance)
//		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
//			return nil, reconcile.Result{}, err
//		}
//		err := r.Create(context.TODO(), pod)
//		if err != nil {
//
//			return nil, reconcile.Result{}, err
//		}
//		return nil, reconcile.Result{Requeue: true}, nil
//	}
//	return &pods.Items[0], reconcile.Result{}, nil
//}
//
//func newPodForCR(cr *dbv1.Mysql) *corev1.Pod {
//	labels := map[string]string{
//		"app":     cr.Name,
//		"version": cr.Spec.Version,
//	}
//	return &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:         cr.Name + "-pod",
//			GenerateName: cr.Name + "-pod",
//			Namespace:    cr.Namespace,
//			Labels:       labels,
//		},
//		Spec: corev1.PodSpec{
//			Containers: []corev1.Container{
//				{
//					Name:  "mysql",
//					Image: "mysql:5.7",
//					Env: []corev1.EnvVar{
//						{
//							Name:  "MYSQL_ROOT_PASSWORD",
//							Value: cr.Spec.Password,
//						},
//					},
//				},
//			},
//		},
//	}
//}
//
//func (r *MysqlReconciler) deleteExternalResources(instance *dbv1.Mysql) error {
//	r.Log.Info("Deleting external resource. Sleep 30 seconds", "name:", instance.Name)
//	time.Sleep(time.Second * 30)
//	return nil
//}
//
//func containsString(slice []string, s string) bool {
//	for _, item := range slice {
//		if item == s {
//			return true
//		}
//	}
//	return false
//}
//
//func removeString(slice []string, s string) (result []string) {
//	for _, item := range slice {
//		if item == s {
//			continue
//		}
//		result = append(result, item)
//	}
//	return
//}
