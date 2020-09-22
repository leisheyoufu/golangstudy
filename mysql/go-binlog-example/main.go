package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/leisheyoufu/golangstudy/mysql/go-binlog-example/pkg"
)

var host = flag.String("host", "127.0.0.1", "MySQL host")
var port = flag.Int("port", 3306, "MySQL port")
var user = flag.String("user", "root", "MySQL user, must have replication privilege")
var password = flag.String("password", "123456", "MySQL password")

func main() {
	flag.Parse()
	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	go pkg.BinlogListener(endpoint, *user, *password)

	time.Sleep(2 * time.Minute)
	fmt.Print("Thx for watching, goodbye")
}
