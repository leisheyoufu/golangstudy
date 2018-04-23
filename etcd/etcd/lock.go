package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	wg        sync.WaitGroup
	lockCount uint32 = 0
)

func aquireLock(cli *clientv3.Client, lockKey string) (*concurrency.Mutex, error) {
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return nil, err
	}
	m := concurrency.NewMutex(session, lockKey)
	if err = m.Lock(context.TODO()); err != nil {
		return nil, err
	}
	return m, nil
}

func releaseLock(cli *clientv3.Client, m *concurrency.Mutex) error {
	// cl: unlock would help delete the key
	return m.Unlock(context.TODO())
}

func lockProc(cli *clientv3.Client, key string, lockErr *error) error {
	defer wg.Done()

	m, err := aquireLock(cli, key)
	if err != nil {
		return err
	}
	swapped := atomic.CompareAndSwapUint32(&lockCount, 0, lockCount+1)
	if !swapped {
		fmt.Println("Lock failed")
		*lockErr = errors.New("Lock failed")
	}
	time.Sleep(1 * time.Second)
	swapped = atomic.CompareAndSwapUint32(&lockCount, 1, lockCount-1)
	if !swapped {
		fmt.Println("Unlock failed")
		*lockErr = errors.New("Unlock failed")
	}
	releaseLock(cli, m)
	return nil
}

func TestLock() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	key := "/service/lock/" + hostname
	wg.Add(4)
	i := 0
	var lockErr error
	for i < 4 {
		go lockProc(cli, key, &lockErr)
		i++
	}
	wg.Wait()
	if lockErr != nil {
		fmt.Printf("TestLock: failed\n")
		return
	}
	if err = delete(cli, key); err != nil {
		fmt.Printf("Dellock: Failed")
		return
	}
	fmt.Printf("TestLock: ok\n")
}
