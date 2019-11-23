package client

import (
	"fmt"
	"reflect"

	"github.com/garyburd/redigo/redis"
)

func ListPushPop(conn redis.Conn) {
	_, err := conn.Do("LPUSH", "list1", "ele1", "ele2", "ele3")
	if err != nil {
		log.Error(err)
		return
	}
	// redis.Values here
	vals, err := redis.Values(conn.Do("LRANGE", "list1", "0", "1"))
	if err != nil {
		log.Error(err)
		return
	}
	for _, val := range vals {
		log.Info(fmt.Sprintf("LRANGE %+v", string(val.([]byte))))
	}
	_, err = conn.Do("LPUSH", "list1", "ele4", "ele5", "ele6")
	if err != nil {
		log.Error(err)
		return
	}
	res, err := redis.String(conn.Do("LPOP", "list1"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("LPOP type %s val %s", reflect.TypeOf(res), res))
	_, err = conn.Do("expire", "list1", 3)
	if err != nil {
		log.Error(err)
		return
	}
}
