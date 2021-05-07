package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"golang.org/x/time/rate"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("example")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

const (
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
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

func WriteFile2(content string) {
	fd, _ := os.OpenFile("/tmp/output2.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd.Write([]byte(content))
	fd.Close()
}

func main() {
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)
	ctx := context.Background()
	// r := rate.Every(1024)
	limit := rate.NewLimiter(1024, 1024*1024)
	for {
		// if limit.AllowN(time.Now(), 8) {
		// 	log.Info("log:event happen")
		// } else {
		// 	log.Info("log:event not allow")
		// 	time.Sleep(1 * time.Second)
		// }
		limit.WaitN(ctx, 1024)
		WriteFile2(GenRandomString(1024))
		log.Info("Write 1024 byte into file")
	}
}
