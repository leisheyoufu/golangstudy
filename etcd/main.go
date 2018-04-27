// study how to use etcd
package main

import (
	"github.com/leisheyoufu/golangstudy/etcd/etcd"
)

func main() {
	etcd.TestSingle()
	etcd.TestBulk()
	etcd.TestTxn()
	etcd.TestWatch()
	etcd.TestLock()
	etcd.TestLease()
}
