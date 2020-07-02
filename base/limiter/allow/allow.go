package main

import (
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

func main() {
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)

	r := rate.Every(1) // 1秒1个

	//r = rate.Limit(1)  // CL: 每秒装r个令牌，最多10个令牌
	limit := rate.NewLimiter(r, 10)
	for {
		if limit.AllowN(time.Now(), 8) {
			log.Info("log:event happen")
		} else {
			log.Info("log:event not allow")
		}

	}

}
