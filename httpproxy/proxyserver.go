package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

//redirect request to http://127.0.0.1:2003
func helloHandler(w http.ResponseWriter, r *http.Request) {

	trueServer := "http://127.0.0.1:2003"

	url, err := url.Parse(trueServer)
	if err != nil {
		log.Println(err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)
	log.Println("proxy handler exit")
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":2002", nil))
}
