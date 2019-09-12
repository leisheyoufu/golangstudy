package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/leisheyoufu/golangstudy/redis/common"
)

const (
	Address = "10.190.177.19:6379"
)

var (
	log  *common.Logger
	Pool redis.Pool
)

func init() {
	Pool = redis.Pool{
		MaxIdle:     16,
		MaxActive:   32,
		IdleTimeout: 120,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", Address)
		},
	}
}

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
	fmt.Printf("Got name: %s \n", name)
	_, err = conn.Do("DEL", "name")
	if err != nil {
		log.Error(err)
		return
	}
	name, err = redis.String(conn.Do("GET", "name"))
	if err != nil {
		log.Error(err)
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
	fmt.Printf("Redis got name: %s", name)
}

func MsetMget(conn redis.Conn) {
	_, err := conn.Do("MSET", "name", "loch", "age", 22)
	if err != nil {
		log.Error(err)
		return
	}
	res, err := redis.Strings(conn.Do("MGET", "name", "age"))
	if err != nil {
		log.Error(err)
		return
	}

	res_type := reflect.TypeOf(res)
	fmt.Printf("res type : %s \n", res_type)
	fmt.Printf("MGET name: %s \n", res)
	fmt.Println(len(res))
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

func ListPushPop(conn redis.Conn) {
	_, err := conn.Do("LPUSH", "list1", "ele1", "ele2", "ele3")
	if err != nil {
		log.Error(err)
		return
	}
	res, err := redis.String(conn.Do("LPOP", "list1"))
	if err != nil {
		log.Error(err)
		return
	}
	res_type := reflect.TypeOf(res)
	fmt.Printf("res type : %s \n", res_type)
	fmt.Printf("res  : %s \n", res)
	_, err = conn.Do("expire", "list1", 3)
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

func Pipelining(conn redis.Conn) {
	conn.Send("HSET", "student", "name", "loch", "age", "22")
	conn.Send("HSET", "student", "Score", "100")
	conn.Send("HGET", "student", "age")
	conn.Flush()

	res1, err := conn.Receive()
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Printf("Receive res1:%v \n", res1)
	res2, err := conn.Receive()
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Printf("Receive res2:%v\n", res2)
	res3, err := conn.Receive()
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Printf("Receive res3:%s\n", res3)
	_, err = conn.Do("expire", "student", 3)
	if err != nil {
		log.Error(err)
		return
	}
}

func Subs() {
	conn, err := redis.Dial("tcp", Address)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()
	psc := redis.PubSubConn{conn}
	err = psc.Subscribe("channel1")
	defer func() {
		psc.Unsubscribe("channel1")
	}()

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			fmt.Println(v)
			return
		}
	}
}

func Push(message string) {
	conn, _ := redis.Dial("tcp", Address)
	defer conn.Close()
	_, err1 := conn.Do("PUBLISH", "channel1", message)
	if err1 != nil {
		log.Error(err1)
		return
	}
}

func PubSub() {
	go Subs()
	go Push("Hello, this is chenglch")
	go Push("Hello, this is loch")

	time.Sleep(time.Second * 10)
}

func MULTI(conn redis.Conn) {
	conn.Send("MULTI")
	conn.Send("INCR", "foo")
	conn.Send("INCR", "bar")
	r, err := conn.Do("EXEC")
	if err != nil {
		log.Error(err)
	}
	fmt.Println(r)
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

func ConnPool() {
	conn := Pool.Get()
	res, err := conn.Do("HSET", "student", "name", "jack")
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Println(res, err)
	res1, err := redis.String(conn.Do("HGET", "student", "name"))
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Printf("res:%s,error:%v", res1, err)
	_, err = conn.Do("DEL", "student")
	if err != nil {
		log.Error(err)
		return
	}
}

func main() {
	common.InitLogger()
	log = common.GetLogger("")
	conn, err := redis.Dial("tcp", Address)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()
	//GetSet(conn)
	//Expire(conn)
	//MsetMget(conn)
	//ListPushPop(conn)
	//HashSetGet(conn)
	//Pipelining(conn)
	//PubSub()
	//MULTI(conn)
	ConnPool()
}
