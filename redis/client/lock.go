package client

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	LockKey   = "redis"
	LockToken = "token"
)

type Lock struct {
	resource string
	token    string
	conn     redis.Conn
	timeout  int
}

func (lock *Lock) tryLock() (ok bool, err error) {
	_, err = redis.String(lock.conn.Do("SET", lock.key(), lock.token, "EX", int(lock.timeout), "NX"))
	if err == redis.ErrNil {
		// The lock was not successful, it already exists.
		log.Info("SETNX key already exist")
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (lock *Lock) Unlock() (err error) {
	_, err = lock.conn.Do("del", lock.key())
	return
}

func (lock *Lock) key() string {
	return fmt.Sprintf("redislock:%s", lock.resource)
}

func (lock *Lock) AddTimeout(ex_time int64) (ok bool, err error) {
	ttl_time, err := redis.Int64(lock.conn.Do("TTL", lock.key()))
	log.Info(fmt.Sprintf("SETNX TTL %d", ttl_time))
	if err != nil {
		log.Error("redis get failed")
	}
	if ttl_time > 0 {
		_, err := redis.String(lock.conn.Do("SET", lock.key(), lock.token, "EX", int(ttl_time+ex_time)))
		if err == redis.ErrNil {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

func TryLock(conn redis.Conn, resource string, token string, DefaulTimeout int) (lock *Lock, ok bool, err error) {
	return TryLockWithTimeout(conn, resource, token, DefaulTimeout)
}

func TryLockWithTimeout(conn redis.Conn, resource string, token string, timeout int) (lock *Lock, ok bool, err error) {
	lock = &Lock{resource, token, conn, timeout}

	ok, err = lock.tryLock()

	if !ok || err != nil {
		lock = nil
	}

	return
}

func LockTask1(timeout int) (lock *Lock) {
	conn, err := redis.Dial("tcp", Address)
	if err != nil {
		log.Error("SETNX Task1 can not connecet to redis")
	}
	lock, ok, err := TryLock(conn, LockKey, LockToken, int(timeout))
	if err != nil {
		log.Error("SETNX Task1 Error while attempting lock")
	}
	if !ok {
		log.Info("SETNX Task1 can not get lock")
		return lock
	}
	log.Info("SETNX Task1 Got lock")
	return lock
}

func LockTask2(timeout int, ch chan<- interface{}) (lock *Lock) {
	defer func() {
		ch <- struct{}{}
	}()
	conn, err := redis.Dial("tcp", Address)
	if err != nil {
		log.Error("SETNX Task2 can not connecet to redis")
	}
	lock, ok, err := TryLock(conn, LockKey, LockToken, int(timeout))
	if err != nil {
		log.Error("SETNX Task2 Error while attempting lock")
	}
	if !ok {
		log.Info("SETNX Task2 can not get lock")
		return lock
	}
	log.Info("SETNX Task2 Got lock")
	return lock
}

func TestLock() {
	DefaultTimeout := 20
	lock := LockTask1(DefaultTimeout / 2)
	time.Sleep(time.Duration(1) * time.Second)
	ch := make(chan interface{}, 1)
	go LockTask2(DefaultTimeout/2, ch)
	<-ch
	lock.Unlock()
	lock = LockTask2(DefaultTimeout/2, ch)
	<-ch
	defer lock.Unlock()
}
