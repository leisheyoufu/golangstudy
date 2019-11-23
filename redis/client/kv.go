package client

import (
	"fmt"
	"reflect"

	"github.com/garyburd/redigo/redis"
)

func GetSet(conn redis.Conn) {
	_, err := conn.Do("SET", "name", "loch")
	if err != nil {
		log.Error(err)
		return
	}
	name, err := redis.String(conn.Do("GET", "name"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("Got val: %s", name))
	_, err = conn.Do("DEL", "name")
	if err != nil {
		log.Error(err)
		return
	}
	name, err = redis.String(conn.Do("GET", "name"))
	if err != nil {
		log.Info(fmt.Sprintf("Expected error %+v", err))
		return
	}
}

func Expire(conn redis.Conn) {
	_, err := conn.Do("SET", "name", "expire value")
	if err != nil {
		log.Error(err)
		return
	}
	_, err = conn.Do("expire", "name", 3)
	if err != nil {
		log.Error(err)
		return
	}
	name, err := redis.String(conn.Do("GET", "name"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("Redis got name: %s", name))
}

func MsetMget(conn redis.Conn) {
	_, err := conn.Do("MSET", "name", "loch", "age", 22)
	if err != nil {
		log.Error(err)
		return
	}
	// Note(loch): []interface{} is needed for MGET keys
	keys := make([]interface{}, 2)
	keys[0] = "loch"
	keys[1] = "age"
	res, err := redis.Strings(conn.Do("MGET", keys...))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("MGET val1: %s", res[0]))
	log.Info(fmt.Sprintf("MGET val2: %s", res[1]))
	// res_type := reflect.TypeOf(res)
	// fmt.Printf("res type : %s \n", res_type)
	// fmt.Printf("MGET name: %s \n", res)
	_, err = conn.Do("expire", "name", 3)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = conn.Do("expire", "age", 3)
	if err != nil {
		log.Error(err)
		return
	}
}

func HashSetGet(conn redis.Conn) {
	_, err := conn.Do("HSET", "student", "name", "loch", "age", 22)
	if err != nil {
		log.Error(err)
		return
	}
	age, err := redis.Int64(conn.Do("HGET", "student", "age"))
	if err != nil {
		log.Error(err)
		return
	}
	res_type := reflect.TypeOf(age)
	fmt.Printf("age type : %s \n", res_type)
	fmt.Printf("age  : %d \n", age)
	name, err := redis.String(conn.Do("HGET", "student", "name"))
	if err != nil {
		log.Error(err)
		return
	}
	res_type = reflect.TypeOf(name)
	fmt.Printf("name type : %s \n", res_type)
	fmt.Printf("name  : %d \n", name)
	_, err = conn.Do("expire", "student", 3)
	if err != nil {
		log.Error(err)
		return
	}
}
