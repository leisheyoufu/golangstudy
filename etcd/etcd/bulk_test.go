package etcd

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
)

func Test_BulkPutDelete_1(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	m := genData(0, bulkNodeNumber-1)
	err = bulkPut(cli, m)
	if err != nil {
		t.Error(err)
		return
	}
	newM, err := bulkQeury(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
	if err != nil {
		t.Error(err)
		return
	}

	if len(newM) != len(m) {
		t.Error("TestBulk: error data len")
		return
	}
	for index, item := range newM {
		if _, ok := m[index]; !ok {
			t.Error("TestBulk: error data index")
			return
		}
		for k, v := range item {
			if _, ok := m[index][k]; !ok {
				t.Error("TestBulk: error data index_k")
				return
			}
			if v != m[index][k] {
				t.Error("TestBulk: error data index_k_v")
				return
			}
		}
	}
	err = bulkDelete(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log("bulk put bulk query delete pass")
}

/*
bulkNodeNumber = 100
go test -test.bench=".*" -count 2
Benchmark_BulkPut_Count100-4    	     100       	  10175775 ns/op
Benchmark_BulkPut_Count100-4    	     100       	  10221784 ns/op

average 10ms to invoke bulkPut

As etcd only support 128 keys in one transaction , so put 5000 nodes need 5 seconds
*/
func Benchmark_BulkPut_Count100(b *testing.B) {
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
		b.StartTimer()
		err = bulkPut(cli, m)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
		err = bulkDelete(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

/*
bulkNodeNumber = 100
go test -test.bench=".*" -count 2
Benchmark_Query_Count100-4      	    1000       	   1315235 ns/op
Benchmark_Query_Count100-4      	    1000       	   1265400 ns/op

1 ms
*/

func Benchmark_BulkQuery_Count100(b *testing.B) {
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
	err = bulkPut(cli, m)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err = bulkQeury(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
	}
	err = bulkDelete(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
	if err != nil {
		fmt.Println(err)
		return
	}
}

/*
bulkNodeNumber = 100
go test -test.bench=".*" -count 2
Benchmark_Delete_Cout100-4     	     100       	  10393542 ns/op
Benchmark_Delete_Count100-4     	     100       	  10677803 ns/op

10 ms
*/

func Benchmark_BulkDelete_Count100(b *testing.B) {
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
		err = bulkPut(cli, m)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StartTimer()
		err = bulkDelete(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
	}
}
