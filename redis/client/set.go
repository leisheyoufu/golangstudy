package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

func Set(conn redis.Conn) {
	elems := [...]string{
		"set",
		"mysql",
		"redis",
		"tidb",
		"hbase",
		"memcached",
		"postgress",
		"db2",
		"oracle",
	}
	dbs := make([]interface{}, len(elems))
	for i, elem := range elems {
		dbs[i] = elem
	}
	_, err := conn.Do("sadd", dbs...)
	if err != nil {
		log.Error(err)
		return
	}
	vals, err := redis.Values(conn.Do("smembers", "set"))
	if err != nil {
		log.Error(err)
		return
	}
	strs := make([]string, 0)
	for _, val := range vals {
		strs = append(strs, string(val.([]byte)))
	}
	log.Info(fmt.Sprintf("Set %s", strings.Join(strs, ",")))
	log.Info(fmt.Sprintf("Set len=%d", len(vals)))
	exist, err := redis.Bool(conn.Do("SISMEMBER", "set", "mysql"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(strconv.FormatBool(exist))
	_, err = conn.Do("SREM", "set", "mysql")
	if err != nil {
		log.Error(err)
		return
	}
	vals, err = redis.Values(conn.Do("smembers", "set"))
	if err != nil {
		log.Error(err)
		return
	}
	for _, val := range vals {
		strs = append(strs, string(val.([]byte)))
	}
	log.Info(fmt.Sprintf("Set %s", strings.Join(strs, ",")))
	log.Info(fmt.Sprintf("Set len=%d", len(vals)))
	_, err = conn.Do("DEL", "set")
	if err != nil {
		log.Error(err)
		return
	}
}

// sorted set
func Zset(conn redis.Conn) {
	scores := [...]int{1, 2, 3, 3, 5}
	words := [...]string{"one", "two", "three", "three", "five"}
	for i := 0; i < 5; i++ {
		_, err := conn.Do("zadd", "zset", scores[i], words[i])
		if err != nil {
			log.Error(err)
			return
		}
	}
	vals, err := redis.Values(conn.Do("zrange", "zset", 0, 10, "withscores"))
	if err != nil {
		log.Error(err)
		return
	}
	strs := make([]string, 0)
	for _, val := range vals {
		strs = append(strs, string(val.([]byte)))
	}
	log.Info(strings.Join(strs, ","))
	_, err = conn.Do("DEL", "zset")
	if err != nil {
		log.Error(err)
		return
	}
}
