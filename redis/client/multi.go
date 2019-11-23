package client

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

func MULTI(conn redis.Conn) {
	conn.Send("MULTI")
	conn.Send("INCR", "foo")
	conn.Send("INCR", "bar")
	conn.Send("INCR", "bar")
	_, err := conn.Do("EXEC")
	if err != nil {
		log.Error(err)
		return
	}
	val, err := redis.Int64(conn.Do("GET", "foo"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("MULTI val1 %d", val))
	val, err = redis.Int64(conn.Do("GET", "bar"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("MULTI val1 %d", val))
	_, err = conn.Do("DEL", "foo")
	if err != nil {
		log.Error(err)
		return
	}
	_, err = conn.Do("DEL", "bar")
	if err != nil {
		log.Error(err)
		return
	}
}
