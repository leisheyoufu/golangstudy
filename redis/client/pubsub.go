package client

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

func Subs(exit <-chan struct{}, ch chan<- struct{}, count int) {
	conn, err := redis.Dial("tcp", Address)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()
	defer log.Info("Sub exit")
	psc := redis.PubSubConn{Conn: conn}
	err = psc.Subscribe("channel1")
	if err != nil {
		ch <- struct{}{}
		log.Error(err)
		return
	}
	ch <- struct{}{}
	defer func() {
		psc.Unsubscribe("channel1")
	}()

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Info(fmt.Sprintf("Sub %s: message: %s\n", v.Channel, v.Data))
			count--
		case redis.Subscription:
			log.Info(fmt.Sprintf("Sub %s: %s %d\n", v.Channel, v.Kind, v.Count))
		case error:
			log.Error(v)
			return
		}
		if count == 0 {
			<-exit
			return
		}
		select {
		case <-exit:
			return
		default:
		}

	}
}

func Push(message string, ch chan<- struct{}) {
	conn, _ := redis.Dial("tcp", Address)
	defer conn.Close()
	_, err1 := conn.Do("PUBLISH", "channel1", message)
	log.Info(fmt.Sprintf("Pub message %s", message))
	if err1 != nil {
		log.Error(err1)
		return
	}
	ch <- struct{}{}
}

func PubSub() {
	ch := make(chan struct{})
	exit := make(chan struct{})
	// NOTE(loch): Add subscriber at first, then add publisher, otherwise message in channel may be dropped.
	go Subs(exit, ch, 2)
	<-ch
	go Push("Hello, this is chenglch", ch)
	go Push("Hello, this is loch", ch)
	<-ch
	<-ch
	exit <- struct{}{}
	close(exit)
	close(ch)
}
