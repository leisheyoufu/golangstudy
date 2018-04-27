package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"runtime"
	"testing"
)

/* If data changed, exit immediately */
func watchExit(cli *clientv3.Client, path string, c chan<- struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	rch := cli.Watch(ctx, path, clientv3.WithPrefix())
	c <- struct{}{}
	for wresp := range rch {
		// NOTE(cl): one operation may contains multiple event
		for range wresp.Events {
			//fmt.Println(string(ev.Kv.Key))
		}
		c <- struct{}{}
		cancel()
	}
}

/*
go test -test.bench="Benchmark_Watch_Count1" -count 2
Benchmark_Watch_Count1-4       	     200       	   9580134 ns/op
Benchmark_Watch_Count1-4       	     100       	  10184329 ns/op

10 ms, nearly no time consumed in watch function
*/
func Benchmark_Watch_Count1(b *testing.B) {
	b.StopTimer()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()
	for i := 0; i < b.N; i++ {
		c := make(chan struct{}, 0)
		go watchExit(cli, "/nodes", c)
		runtime.Gosched()
		<-c
		b.StartTimer()
		err = put(cli, testKey, testData)
		if err != nil {
			fmt.Println(err)
			return
		}
		<-c
		b.StopTimer()
		err = delete(cli, testKey)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

/*
go test -test.bench="Benchmark_Watch_Count100" -count 2
Benchmark_Watch_Count100-4     	     100       	  10057483 ns/op
Benchmark_Watch_Count100-4     	     200       	  10029685 ns/op
*/
func Benchmark_Watch_Count100(b *testing.B) {
	b.StopTimer()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()
	m := genData(0, bulkNodeNumber-1)
	for i := 0; i < b.N; i++ {
		c := make(chan struct{}, 0)
		go watchExit(cli, "/nodes", c)
		runtime.Gosched()
		<-c
		b.StartTimer()
		err = bulkPut(cli, m)
		if err != nil {
			fmt.Println(err)
			return
		}
		<-c
		b.StopTimer()
		err = bulkDelete(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
