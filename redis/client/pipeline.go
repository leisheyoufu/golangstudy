package client

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

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
	log.Info(fmt.Sprintf("Pipeline Recive1 %+v", res1)) // Recive1 2 : two key
	res2, err := conn.Receive()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("Pipeline Recive2 %+v", res2)) // Recive1 1 : one key
	res3, err := conn.Receive()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(fmt.Sprintf("Pipeline Recive3 %+v", string(res3.([]byte)))) // Recive3: 22
	_, err = conn.Do("expire", "student", 3)
	if err != nil {
		log.Error(err)
		return
	}
}
