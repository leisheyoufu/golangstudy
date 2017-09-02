package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leisheyoufu/golangstudy/rest/api"
)

func main() {
	api.Router = mux.NewRouter().StrictSlash(true)
	api.NewNodeApi(api.Router)
	log.Fatal(http.ListenAndServe(":8080", api.Router))
}
