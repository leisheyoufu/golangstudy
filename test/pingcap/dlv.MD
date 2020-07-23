## Setup delve
go get github.com/go-delve/delve/cmd/dlv

## pingcap dm unit test
github.com/pingcap/dm/pkg/binlog/event/common_test.go
```
func (t *testCommonSuite) TestGenCommonGTIDEvent(c *C) {
	var (
		flavor           = gmysql.MySQLFlavor
		serverID  uint32 = 101
		gSetStr          = ""
		gSet      gtid.Set
		latestPos uint32 = 123
	)

	// nil gSet, invalid
	gtidEv, err := GenCommonGTIDEvent(flavor, serverID, latestPos, gSet)
	c.Assert(err, NotNil)
	c.Assert(gtidEv, IsNil)
```
go test -check.f "testCommonSuit" -check.list

## pingcap dm unit test with dlv
```
(dlv) b github.com/pingcap/dm/pkg/binlog/event/common_test.go:43
Breakpoint 1 set at 0x1588ce9 for github.com/pingcap/dm/pkg/binlog/event.(*testCommonSuite).TestGenCommonFileHeader() ./common_test.go:43
(dlv) r
Process restarted with PID 79893
(dlv) c
github.com/pingcap/dm/pkg/binlog/event.(*testCommonSuite).TestGenCommonFileHeader() ./common_test.go:43 (hits goroutine(20):1 total:1) (PC: 0x1588ce9)
```

dlv test github.com/pingcap/dm/pkg/binlog/event/
(dlv) b github.com/pingcap/dm/pkg/binlog/event/generator_test.go:128
(dlv) b github.com/pingcap/dm/pkg/binlog/event/generator_test.go:304


## set string lenth to avoid of truncate
(dlv) config -list
...
max-string-len         1000
max-array-values       1000
...
(dlv) config max-string-len 1000