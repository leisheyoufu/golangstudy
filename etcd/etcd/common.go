package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"runtime"
	"strings"
	"time"
)

const (
	bulkNodeNumber = 100
	requestTimeout = 5 * time.Second
	dialTimeout    = 5 * time.Second
	debug          = false
	testKey        = "/nodes/node1"
	testData       = `{"driver":"ssh", "params": {"port":22, "user":"root"}`
)

var (
	endpoints = [...]string{"127.0.0.1:2379"}
)

func notFound(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.ModRevision(key), "=", 0)
}

func found(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.ModRevision(key), "!=", 0)
}

func contextError(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("Timeout")
	}
	return ctx.Err()
}

func printMsg(msg string) {
	if !debug {
		return
	}
	fmt.Printf(msg)
}

func printFunc() {
	if !debug {
		return
	}
	pc, file, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	paths := strings.Split(file, "/")
	if ok {
		fmt.Printf("%s:%d %s called\n", paths[len(paths)-1], line, f.Name())
	}
}
