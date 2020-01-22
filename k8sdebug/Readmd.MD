## Debug
$env:GO111MODULE="on"  // powershell
$env:GOPROXY="https://goproxy.cn"  // powershell

go env -w GOPROXY=https://goproxy.cn,direct  // linux 

go run main.go <pod> -n loch

## Dlv
dlv debug main.go -- httpd-6dd946dcdc-dpdth -n loch


## Reference
\k8s.io\kubernetes\pkg\kubectl\cmd\cmd.go    NewKubectlCommand
staging\src\k8s.io\kubectl\pkg\cmd\exec\exec.go   NewCmdExec