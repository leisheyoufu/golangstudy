package main

import (
	"log"
	"os"

	"github.com/leisheyoufu/golangstudy/ssh/common"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <host:port> <user> <password>", os.Args[0])
	}
	sshInst := common.NewPasswordSSH(os.Args[1], os.Args[2], os.Args[3])
	err := sshInst.ConnectToHost()
	if err != nil {
		panic(err)
	}
	console, err := sshInst.StartConsole(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
	console.Start("console.log")
}
