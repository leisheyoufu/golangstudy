package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func validate(index int, events []*clientv3.Event, cancel func()) (err error) {
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	switch index {
	case 0:
		for _, ev := range events {
			if ev.Type != 0 || string(ev.Kv.Key) != testKey || string(ev.Kv.Value) != testData {
				return
			}
		}
	case 1:
		newM := genData(0, 3)
		for _, ev := range events {
			var ok bool
			if ev.Type != 0 {
				return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
			}
			if _, ok = newM[string(ev.Kv.Key)]; !ok {
				return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
			}
			var m map[string]string
			if err := json.Unmarshal(ev.Kv.Value, &m); err != nil {
				return err
			}
			for k, _ := range m {
				if _, ok := newM[string(ev.Kv.Key)][k]; !ok {
					return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
				}
				if newM[string(ev.Kv.Key)][k] != m[k] {
					return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
				}
			}
		}
	case 2:
		paths := [...]string{"/nodes/bulk/0", "/nodes/bulk/1", "/nodes/bulk/2", "/nodes/bulk/3"}
		for _, ev := range events {
			if ev.Type != 1 {
				return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
			}
			i := 0
			for i < len(paths) {
				if paths[i] == string(ev.Kv.Key) {
					break
				}
				i++
			}
			if i == len(paths) {
				return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
			}
		}
	case 3:
		for _, ev := range events {
			if ev.Type != 1 {
				return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
			}
			if string(ev.Kv.Key) != "/nodes/bulk/3" {
				return errors.New(fmt.Sprintf("Failed to validate watch event %d", index))
			}
			cancel()
		}
	}
	return nil

}

func watch(cli *clientv3.Client, path string) {
	ctx, cancel := context.WithCancel(context.Background())
	rch := cli.Watch(ctx, path, clientv3.WithPrefix())
	i := 0
	// NOTE(cl): one operation may generate data in channel
	for wresp := range rch {
		// NOTE(cl): one operation may contains multiple event
		if err := validate(i, wresp.Events, cancel); err != nil {
			fmt.Println(err)
			return
		}
		i++
		// cancel() // If do not call cancel here, watch can not be stopoed
	}
	cancel()
}

func dataChange(cli *clientv3.Client) {
	time.Sleep(time.Duration(1) * time.Second)
	var err error
	err = put(cli, testKey, testData)
	if err != nil {
		fmt.Println(err)
		return
	}
	m := genData(0, 3)
	err = bulkPut(cli, m)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = bulkDelete(cli, "/nodes/bulk", 0, 2)
	if err != nil {
		fmt.Println(err)
		return
	}
	// NOTE(cl): if the key deleted is not exist, no event will be generated in the watch statement
	err = delete(cli, "/nodes/bulk/3")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func TestWatch() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()
	// watch single put
	go dataChange(cli)
	watch(cli, "/nodes")
	err = delete(cli, testKey)
	if err != nil {
		fmt.Printf("Failed to delete key %s\n", testKey)
		return
	}
	fmt.Printf("TestWatch: ok\n")
}
