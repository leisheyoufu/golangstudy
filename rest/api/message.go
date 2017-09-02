package api

import (
	"log"
	"net/http"
	"time"
)

func handle(w http.ResponseWriter, req *http.Request, code int, err error) {
	start := time.Now()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		log.Printf("%d\t%s\t%s\t%s\t%s", code, req.Method, req.RequestURI, time.Since(start), err)
		w.WriteHeader(code)
	} else {
		log.Printf("%d\t%s\t%s\t%s", code, req.Method, req.RequestURI, time.Since(start))
		w.WriteHeader(http.StatusOK)
	}
}

func info(req *http.Request) {
	start := time.Now()
	log.Printf("%s\t%s\t%s", req.Method, req.RequestURI, time.Since(start))
}
