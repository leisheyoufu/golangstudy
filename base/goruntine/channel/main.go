package main

import "fmt"

func printA(notify chan struct{}, wait chan struct{}) {
	i := 1
	fmt.Println(i)
	notify <- struct{}{}
	for i < 100 {
		<-wait
		i += 2
		if i >= 100 {
			break
		}
		fmt.Println(i)
		notify <- struct{}{}
	}
}

func printB(notify chan struct{}, wait chan struct{}) {
	i := 0
	for i <= 100 {
		<-notify
		i += 2
		if i > 100 {
			break
		}
		fmt.Println(i)
		wait <- struct{}{}
	}
}

func main() {
	notify := make(chan struct{})
	wait := make(chan struct{})
	go printB(notify, wait)
	printA(notify, wait)
}
