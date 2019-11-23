package main

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/leisheyoufu/golangstudy/redis/client"
)

const (
	Address = "192.168.126.10:6379"
)

func main() {
	conn, err := redis.Dial("tcp", Address)
	if err != nil {
		fmt.Errorf("Can not connect to redis err %v", err)
		return
	}
	defer conn.Close()
	client.GetSet(conn)
	client.Expire(conn)
	client.MsetMget(conn)
	client.ListPushPop(conn)
	client.HashSetGet(conn)
	client.Pipelining(conn)
	client.PubSub()
	client.MULTI(conn)
	client.ConnPool()
	client.TestLock()
	client.Set(conn)
	client.Zset(conn)
}
