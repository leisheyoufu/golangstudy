## Build failpoint
git clone https://github.com/pingcap/failpoint.git
cd failpoint
make
cp failpoint-ctl $GOBIN/

## Enable failpoint
go get -v github.com/pingcap/failpoint@9f69995143ce   // igonore this statement if go.mod has the correct version
failpoint-ctl enable
// failpoint 会找到包的标志，然后更改埋点的代码，制造错误的场景看程序的反应
GO_FAILPOINTS="main/testPanic=return(true)" go run failpoint.go binding__failpoint_binding__.go
failpoint-ctl disable

## Reference
- [failpoint](https://github.com/pingcap/failpoint)
- [golang 故障模拟工具failpoint的使用](https://blog.csdn.net/lanyang123456/article/details/103091671)