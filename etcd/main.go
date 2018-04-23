// study how to use etcd
package main

import (
	"github.com/leisheyoufu/golangstudy/etcd/etcd"
)

const ()

/*
func Wacher() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	log.Println("watch never cancel")
	// Backgroud never cancel
	ctx, cancel := context.WithCancel(context.Background())
	rch := cli.Watch(ctx, "/nodes", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
		cancel() // If do not call cancel here, watch can not be stopoed
	}
}
*/
func main() {
	etcd.TestSingle()
	etcd.TestBulk()
	etcd.TestTxn()
	etcd.TestWatch()
	etcd.TestLock()
	etcd.TestLease()
}
