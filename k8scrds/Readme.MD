This directory reference https://github.com/zq2599/blog_demos/tree/master/k8s_customize_controller and the blog address is https://blog.csdn.net/boling_cavalry/article/details/88917818

## Import crds definition
kubectl apploy -f crd/student.yaml
kubectl get crd

## Generate controller code
mkdir -p pkg/apis/loch
pkg/client is the target directory, if go module is enabled, git add those files at first, then run the script below

```
SCRIPT_ROOT=/opt/golang/pkg/mod/k8s.io/code-generator@v0.17.0
$SCRIPT_ROOT/generate-groups.sh all github.com/leisheyoufu/golangstudy/k8scrds/pkg/client \
          github.com/leisheyoufu/golangstudy/k8scrds/pkg/apis loch:v1 \
          --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt
```

## Go run controller
If there are multiple go files under main package, should use go build main.go controller.go (or go build *.go) to compile the go files, otherwise, undefine function will be reported.
```
kubectl apply -f crd/student.yaml
$env:GO111MODULE="on"
$env:GOPROXY="https://goproxy.cn"
go run main.go controller.go
kubectl apply -f crd/example.yaml
kubectl delete -f crd/example.yaml
```
Result
```
I1222 12:12:15.805881   13064 controller.go:72] Setting up event handlers
I1222 12:12:15.814855   13064 controller.go:96] Start controller service and sync data
I1222 12:12:15.915609   13064 controller.go:101] worker start
I1222 12:12:15.915609   13064 controller.go:106] worker started
I1222 12:12:28.998862   13064 controller.go:181] 这里是student对象的期望状态: &v1.Student{TypeMeta:v1.TypeMeta{Kind:"", APIVersion:""}, ObjectMeta:v1.ObjectMeta{Name:"object-student", GenerateName:"", Namespace:"default", SelfLink:"/apis/loch.k8s.io/v1/namespaces/default/students/object-student", UID:"c34338ba-bc6a-4c9b-9ef4-d08f51637e7e", ResourceVersion:"179710", Generation:1, CreationTimestamp:v1.Time{Time:time.Time{wall:0x0, ext:63712613548, loc:(*time.Location)(0x21082a0)}}, DeletionTimestamp:(*v1.Time)(nil), DeletionGracePeriodSeconds:(*int64)(nil), Labels:map[string]string(nil), Annotations:map[string]string{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"loch.k8s.io/v1\",\"kind\":\"Student\",\"metadata\":{\"annotations\":{},\"name\":\"object-student\",\"namespace\":\"default\"},\"spec\":{\"name\":\"loch\",\"school\":\"scu\"}}\n"}, OwnerReferences:[]v1.OwnerReference(nil), Finalizers:[]string(nil), ClusterName:"", ManagedFields:[]v1.ManagedFieldsEntry(nil)}, Spec:v1.StudentSpec{name:"", school:""}} ...
I1222 12:12:29.000394   13064 controller.go:182] 实际状态是从业务层面得到的，此处应该去的实际状态，与期望状态做对比，并根据差异做出响应(新增或者删除)
I1222 12:12:29.015596   13064 controller.go:145] Successfully synced 'default/object-student'
I1222 12:12:29.015596   13064 event.go:281] Event(v1.ObjectReference{Kind:"Student", Namespace:"default", Name:"object-student", UID:"c34338ba-bc6a-4c9b-9ef4-d08f51637e7e", APIVersion:"loch.k8s.io/v1", ResourceVersion:"179710", FieldPath:""}): type: 'Normal' reason: 'Synced' Student synced successfully
I1222 12:17:12.026350   13064 controller.go:171] Student对象被删除，请在这里执行实际的删除业务: default/object-student ...
I1222 12:17:12.026350   13064 controller.go:145] Successfully synced 'default/object-student'
```