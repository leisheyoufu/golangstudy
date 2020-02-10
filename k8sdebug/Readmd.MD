## Debug
$env:GO111MODULE="on"  // powershell
$env:GOPROXY="https://goproxy.cn"  // powershell

go env -w GOPROXY=https://goproxy.cn,direct  // linux 

go run main.go <pod> -n loch

## Dlv
dlv debug main.go -- httpd-6dd946dcdc-dpdth -n loch
dlv debug main.go -- airflow-deploy-5d8c4f9d9c-lrmsk -n airflow

```
(dlv) b k8sdebug/pkg/client/client.go:76
Breakpoint 1 set at 0x1884162 for github.com/leisheyoufu/golangstudy/k8sdebug/pkg/client.ForkDebugPod() ./pkg/client/client.go:76
```

## Build
make

## Example
go run main.go airflow-deploy-5d8c4f9d9c-lrmsk -n airflow
kubectl debug airflow-deploy-5d8c4f9d9c-lrmsk --config /root/.kube/config -n airflow

## Reference
\k8s.io\kubernetes\pkg\kubectl\cmd\cmd.go    NewKubectlCommand
staging\src\k8s.io\kubectl\pkg\cmd\exec\exec.go   NewCmdExec