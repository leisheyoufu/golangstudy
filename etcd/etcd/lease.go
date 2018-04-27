package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

const (
	serverHeartbeat = 3
	keepAliveKey    = "/service/hosts/etcdtest"
	serviceLock     = "/service/lock/etcdtest"
)

func registerService(cli *clientv3.Client, id string, ready chan<- struct{}) error {
	m, err := aquireLock(cli, serviceLock)
	if err != nil {
		return err
	}
	if ready != nil {
		ready <- struct{}{}
	}
	_, err = get(cli, keepAliveKey)
	if err == nil {
		// has value
		releaseLock(cli, m)
		return errors.New(fmt.Sprintf("%s: Already registered", id))
	}
	lease := clientv3.NewLease(cli)
	resp, err := lease.Grant(context.TODO(), serverHeartbeat)
	if err != nil {
		releaseLock(cli, m)
		return err
	}
	_, err = cli.Put(context.TODO(), keepAliveKey, "", clientv3.WithLease(resp.ID))
	if err != nil {
		releaseLock(cli, m)
		return err
	}
	releaseLock(cli, m)
	t := serverHeartbeat - 2
	count := 0
	for {
		time.Sleep(time.Duration(t) * time.Second)
		printMsg(fmt.Sprintf("%s: Keepalive\n", id))
		_, err = cli.KeepAliveOnce(context.TODO(), resp.ID)
		if err != nil {
			printMsg("Could not send keepalive request")
			return errors.New("Could not send keepalive request")
		}
		if count == 2 {
			printMsg("Stop keepalive groutine")
			break
		}
		count++
	}
	return nil
}

func TestLease() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()
	ready := make(chan struct{})
	go registerService(cli, "master", ready)
	<-ready
	for {
		err = registerService(cli, "slave", nil)
		if err != nil {
			if err.Error() == "slave: Already registered" {
				continue
			}
			break
		}
		break
	}
	fmt.Printf("TestLease: ok\n")
}
