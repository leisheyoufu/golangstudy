package main

import (
	"fmt"
	"go.uber.org/atomic"
	"sync"
)

func main() {
	m := sync.Map{}
	m.Store("aa", atomic.NewInt64(54))
	v, ok := m.Load("aa")
	if ok {
		fmt.Println(v.(*atomic.Int64).Load())
	}
}