package etcd

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
)

const (
	sigleDataCount_100 = 100
	sigleDataCount_500 = 500
)

func Test_PutQueryDelete(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Close()
	err = put(cli, testKey, testData)
	if err != nil {
		t.Error(err)
		return
	}
	value, err := get(cli, testKey)
	if err != nil {
		t.Error(err)
		return
	}
	if string(value) != testData {
		t.Errorf("Error data: %s\n", value)
		return
	}
	err = delete(cli, testKey)
	if err != nil {
		t.Error(err)
		return
	}
}

/*
go test -test.bench=".*" -count 2
Benchmark_Put_Count1-4                	     100       	  10276641 ns/op
Benchmark_Put_Count1-4                	     100       	  10536987 ns/op
10 ms
*/
func Benchmark_Put_Count1(b *testing.B) {
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
		b.StartTimer()
		err = put(cli, testKey, testData)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
		err = delete(cli, testKey)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

/*
go test -test.bench=".*" -count 2
Benchmark_Get_Count1-4                	    5000       	    277575 ns/op
Benchmark_Get_Count1-4                	    5000       	    268572 ns/op

0.2 ms
*/
func Benchmark_Get_Count1(b *testing.B) {
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
	err = put(cli, testKey, testData)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err := get(cli, testKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()

	}
	err = delete(cli, testKey)
	if err != nil {
		fmt.Println(err)
		return
	}
}

/*
go test -test.bench=".*" -count 2
Benchmark_Delete_Count1-4             	     100       	  10437127 ns/op
Benchmark_Delete_Count1-4             	     200       	  10433204 ns/op

10 ms
*/
func Benchmark_Delete_Count1(b *testing.B) {
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
		err = put(cli, testKey, testData)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StartTimer()
		err = delete(cli, testKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
	}
}

/*
go test -test.bench=".*" -count 2
Benchmark_Put-4                	     100       	  10276641 ns/op
Benchmark_Put-4                	     100       	  10536987 ns/op
10 ms
*/
type SingleData struct {
	key  string
	data string
}

func genSingleData(count int) []SingleData {
	d := make([]SingleData, count)
	for i := 0; i < count; i++ {
		d[i] = SingleData{key: fmt.Sprintf("/nodes/node%d", i),
			data: `{"driver":"ssh", "params": {"port":22, "user":"root"}`}
	}
	return d
}

func putOneThread(cli *clientv3.Client, data []SingleData) error {
	var err error
	for _, d := range data {
		err = put(cli, d.key, d.data)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func putMultiThread(cli *clientv3.Client, data []SingleData) error {
	var err error
	wg.Add(len(data))
	for _, d := range data {
		go func(d SingleData) {
			defer wg.Done()
			err = put(cli, d.key, d.data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}(d)
	}
	wg.Wait()
	return nil
}

func delOneThread(cli *clientv3.Client, data []SingleData) error {
	var err error
	for _, d := range data {
		err = delete(cli, d.key)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func delMultiThread(cli *clientv3.Client, data []SingleData) error {
	var err error
	wg.Add(len(data))
	for _, d := range data {
		go func(d SingleData) {
			defer wg.Done()
			err = delete(cli, d.key)
			if err != nil {
				fmt.Println(err)
				return
			}
		}(d)
	}
	wg.Wait()
	return nil
}

/*
Benchmark_PutOneThread_Count100-4      	       1       	1033333447 ns/op
Benchmark_PutOneThread_Count100-4      	       1       	1028240266 ns/op

1s
*/
func Benchmark_PutOneThread_Count100(b *testing.B) {
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
	data := genSingleData(sigleDataCount_100)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err = putOneThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
		err = delOneThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

/*
Benchmark_PutMultiThread_Count100-4    	     100       	  16285331 ns/op
Benchmark_PutMultiThread_Count100-4    	     100       	  14583633 ns/op
15 ms
*/
func Benchmark_PutMultiThread_Count100(b *testing.B) {
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
	data := genSingleData(sigleDataCount_100)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err = putMultiThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
		err = delMultiThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

/*
Benchmark_PutOneThread_Count500-4      	       1       	5229672837 ns/op
Benchmark_PutOneThread_Count500-4      	       1       	5231150834 ns/op
5s
*/
func Benchmark_PutOneThread_Count500(b *testing.B) {
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
	data := genSingleData(sigleDataCount_500)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err = putOneThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
		err = delOneThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

/*
Benchmark_PutMultiThread_Count500-4    	      30       	  47241669 ns/op
Benchmark_PutMultiThread_Count500-4    	      20       	  50490345 ns/op
50 ms
*/
func Benchmark_PutMultiThread_Count500(b *testing.B) {
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
	data := genSingleData(sigleDataCount_500)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err = putMultiThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.StopTimer()
		err = delMultiThread(cli, data)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
