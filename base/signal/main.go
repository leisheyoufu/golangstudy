package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"context"
	"time"
)

type Func func() error

func SetupSignalHandler(ctx context.Context, handler Func, signals... os.Signal) {
	// 监听指定信号 platform stop 会发term 信号
	c := make(chan os.Signal)
	signal.Notify(c, signals...)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGTERM:
				log.Printf( "signal %v received", s)
				handler()
			case syscall.SIGINT:
				log.Printf( "signal %v received", s)
				handler()
			case syscall.SIGHUP:
				log.Printf( "signal %v received", s)
				handler()
			default:
				log.Printf( "signal %v received", s)
			}
		}
	}()
}


func main() {
	SetupSignalHandler(context.Background(), func() error {
		log.Printf("hadler sighup called for sighub")
		return nil
	}, syscall.SIGHUP)
	pid := os.Getpid()
	log.Printf("进程 PID: %d \n", pid)
	SetupSignalHandler(context.Background(), func() error {
		log.Printf("hadler sigint, sigterm called for sighub")
		return nil
	}, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("sleep 5 minutes")
	time.Sleep(5 * time.Minute)
}
