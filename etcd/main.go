// study how to use etcd
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{"127.0.0.1:2379"}
)

func KV() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	log.Println("Enroll node")
	if _, err := cli.Put(context.TODO(), "/nodes/node1", `{"driver":"ssh", "params": {"port":22, "user":"root"}`); err != nil {
		log.Fatal(err)
	}
	log.Println("Get value")
	if resp, err := cli.Get(context.TODO(), "hello"); err != nil {
		log.Fatal(err)
	} else {
		log.Println("resp: ", resp)
	}
	log.Println("transaction")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = cli.Txn(ctx).
		If(clientv3.Compare(clientv3.Value("/nodes/node1"), "=", `{"driver":"ssh", "params": {"port":22, "user":"root"}`)).
		Then(clientv3.OpPut("/nodes/node2", `{"driver":"ssh", "params": {"port":22, "user":"root"}`)).
		Else(clientv3.OpPut("/nodes/node3", `{"driver":"ssh", "params": {"port":22, "user":"root"}`)).
		Commit()
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}

func Wacher() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
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
func main() {
	KV()
	Wacher()
}
