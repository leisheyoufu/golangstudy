package main

import (
    "io"
    "net/http"
    "log"
    "strings"
    "time"
)

func helloHandler(w http.ResponseWriter, req *http.Request) {
        io.WriteString(w, strings.Join([]string{time.Now().Format("2006-01-02 15:04:05"), "hello, world!\n"}, " "))
}

func main() {
        http.HandleFunc("/hello", helloHandler)
        log.Fatal(http.ListenAndServe(":2003", nil))
}