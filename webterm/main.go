// study how to use etcd
package main

import (
	"flag"
	"fmt"
	"github.com/leisheyoufu/golangstudy/webterm/backend/handler"
	"github.com/spf13/pflag"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	argInsecurePort        = pflag.Int("insecure-port", 9090, "The port to listen to for incoming HTTP requests.")
	argInsecureBindAddress = pflag.IP("insecure-bind-address", net.IPv4(127, 0, 0, 1), "The IP address on which to serve the --port (set to 0.0.0.0 for all interfaces).")
)

func main() {
	log.SetOutput(os.Stdout)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	http.Handle("/term", handler.NewWebSocketHandler())
	http.Handle("/", handler.MakeGzipHandler(handler.NewWebHandler()))
	addr := fmt.Sprintf("%s:%d", *argInsecureBindAddress, *argInsecurePort)
	go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()
	select {}
}
