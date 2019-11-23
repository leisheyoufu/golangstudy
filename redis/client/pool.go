package client

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

var (
	pool redis.Pool
)

func init() {
	pool = redis.Pool{
		MaxIdle:     16, // pool size
		MaxActive:   32,
		IdleTimeout: 120,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", Address)
		},
	}
}

func ConnPool() {
	conn := pool.Get()
	count, err := redis.Int64(conn.Do("HSET", "student", "name", "jack"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("HSET reply count %d", count))
	val, err := redis.String(conn.Do("HGET", "student", "name"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("HGET value: %s", val))
	_, err = conn.Do("DEL", "student")
	if err != nil {
		log.Error(err)
		return
	}
}
