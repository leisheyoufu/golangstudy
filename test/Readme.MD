## Useful command
go test -check.list
go test -test.bench=".*"

$ go test -check.f MyTestSuite   //选择测试套件
$ go test -check.f "Test.*Works"  //选择测试方法
$ go test -check.f "MyTestSuite.Test.*Works"  //选择套件中的方法



## Reference
hello_test1
[go-check](https://labix.org/gocheck)
[go-check-source](https://github.com/go-check/check)
[更全面讲解gocheck](https://www.jianshu.com/p/2084fef262a7)

hello_test2
[gocheck框架](https://www.cnblogs.com/dapingguo/p/12084371.html)

