package main

import (
	"fmt"
	"reflect"
	"time"
)

func GetObjectQueue(l interface{}) chan interface{} {
	s := reflect.ValueOf(l)
	c := make(chan interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		c <- s.Index(i).Interface()
	}
	close(c)
	return c
}

func TestGetItemFromCloseQueue() {
	var a = []string{"1", "2", "3"}
	q := GetObjectQueue(a)
	for s := range q {
		fmt.Printf("Get item from close queue %d\n", s)
	}
	fmt.Printf("Get item from close queue %s, should be nil\n", <-q)
}

func SafeClose(ch chan interface{}) {
	defer func() {
		if recover() != nil {
			// close(ch) panic occur
		}
	}()

	close(ch) // panic if ch is closed
}

func IsClosed(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true // channel has value , can not decide whether the channel is closed
	default:
	}

	return false
}

func IsClosed2(ch <-chan interface{}) bool {
	select {
	case _, ok := <-ch:
		if !ok {
			return true // channel has value , can not decide whether the channel is closed
		}
	default:
	}

	return false
}

func TestCloseChannel() {
	c := make(chan interface{}, 2)
	go func(ch chan<- interface{}) {
		ch <- 1
		ch <- 2
	}(c)
	time.Sleep(1 * time.Second)
	fmt.Printf("Test Is close %v\n", IsClosed2(c))
	close(c)
	fmt.Printf("Test Is close %v\n", IsClosed2(c))
	fmt.Printf("Test Is close %v\n", IsClosed(c))
	fmt.Printf("Test Is close %v\n", IsClosed2(c))
	SafeClose(c)
}

func main() {
	TestGetItemFromCloseQueue()
	TestCloseChannel()
}
