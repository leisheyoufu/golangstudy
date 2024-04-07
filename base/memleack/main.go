package main

import (
	"math/rand"
	"time"
)

func GenRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := make([]byte, l)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result[i] = bytes[r.Intn(len(bytes))]
	}
	return string(result)
}

func Memleak(n int) {
	var a []string
	for {
		s := GenRandomString(n * 1024 * 1024)
		a = append(a, s)
		time.Sleep(1 * time.Second)
	}
}
func main() {
	Memleak()
}
