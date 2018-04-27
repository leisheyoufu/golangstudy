package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
)

func bulkPut(cli *clientv3.Client, m map[string]map[string]string) error {
	var err error
	printFunc()
	var response *clientv3.TxnResponse
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	ops := make([]clientv3.Op, len(m))
	i := 0
	for k, v := range m {
		var data []byte
		if data, err = json.Marshal(v); err != nil {
			return err
		}
		ops[i] = clientv3.OpPut(k, string(data))
		i++
	}
	// NOTE(cl): by default the length of ops could not larger than 128 see
	// see https://github.com/coreos/etcd/issues/7826
	response, err = cli.Txn(ctx).Then(ops...).Commit()
	if ctxErr := contextError(ctx); ctxErr != nil {
		return ctxErr
	}
	if err != nil {
		return err
	}
	if !response.Succeeded {
		return errors.New("txn error: bulkPut")
	}
	return nil
}

func bulkQeury(cli *clientv3.Client, path string, start int, end int) (map[string]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	printFunc()
	ops := make([]clientv3.Op, end-start+1)
	for i := start; i <= end; i++ {
		var name string
		name = fmt.Sprintf("%s/%d", path, i)
		ops[i] = clientv3.OpGet(name)
	}
	// NOTE(cl): by default the length of ops could not larger than 128 see
	// see https://github.com/coreos/etcd/issues/7826
	response, err := cli.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return nil, err
	}
	ret := make(map[string]map[string]string)
	// ret := make(map[string]map[string]interface{})
	if response.Succeeded {
		for _, temp := range response.Responses {
			getResp := temp.GetResponseRange()
			if len(getResp.Kvs) == 0 {
				continue
			}
			kv := getResp.Kvs[0]
			var m map[string]string
			if err = json.Unmarshal(kv.Value, &m); err != nil {
				return nil, err
			}
			ret[string(kv.Key)] = m
		}
	}
	return ret, nil
}

func bulkDelete(cli *clientv3.Client, path string, start, end int) error {
	var err error
	printFunc()
	var response *clientv3.TxnResponse
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	printFunc()
	ops := make([]clientv3.Op, end-start+1)
	for i := start; i <= end; i++ {
		var name string
		name = fmt.Sprintf("%s/%d", path, i)
		ops[i] = clientv3.OpDelete(name)
	}
	// NOTE(cl): by default the length of ops could not larger than 128 see
	// see https://github.com/coreos/etcd/issues/7826
	response, err = cli.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return err
	}
	if !response.Succeeded {
		return errors.New("Failed to get response bulkDelete")
	}
	return nil
}

func genData(start, end int) map[string]map[string]string {
	m := make(map[string]map[string]string)
	for i := start; i <= end; i++ {
		value := map[string]string{"driver": "ssh", "port": "22", "user": "root", "id": fmt.Sprintf("%d", i)}
		m[fmt.Sprintf("/nodes/bulk/%d", i)] = value
	}
	return m
}
func TestBulk() {
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
	newM, err := bulkQeury(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(newM) != len(m) {
		fmt.Printf("TestBulk: error data len")
		fmt.Printf("%d %d\n", len(newM), len(m))
		return
	}
	for index, item := range newM {
		if _, ok := m[index]; !ok {
			fmt.Printf("TestBulk: error data index")
			return
		}
		for k, v := range item {
			if _, ok := m[index][k]; !ok {
				fmt.Printf("TestBulk: error data index_k")
				return
			}
			if v != m[index][k] {
				fmt.Printf("TestBulk: error data index_k_v")
				return
			}
		}
	}
	err = bulkDelete(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
	if err != nil {
		fmt.Println(err)
		return
	}
	newM, err = bulkQeury(cli, "/nodes/bulk", 0, bulkNodeNumber-1)
	fmt.Printf("TestBulk: ok\n")
}
