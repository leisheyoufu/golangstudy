package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"strings"
)

func get(cli *clientv3.Client, key string) ([]byte, error) {
	var err error
	var resp *clientv3.GetResponse
	printFunc()
	if resp, err = cli.Get(context.TODO(), key); err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, errors.New(fmt.Sprintf("Key %s: not found", key))
	}
	return resp.Kvs[0].Value, nil
}

func put(cli *clientv3.Client, key string, value string) error {
	var err error
	printFunc()
	if _, err = cli.Put(context.TODO(), key, value); err != nil {
		return err
	}
	return nil
}

func delete(cli *clientv3.Client, key string) error {
	var err error
	printFunc()
	if _, err = cli.Delete(context.TODO(), key); err != nil {
		return err
	}
	return nil
}

func TestSingle() {
	var err error
	var cli *clientv3.Client
	cli, err = clientv3.New(clientv3.Config{
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
	value, err := get(cli, testKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	if string(value) != testData {
		fmt.Printf("Error data: %s\n", value)
		return
	}
	err = delete(cli, testKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	value, err = get(cli, testKey)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("TestSingle: ok\n")
}
