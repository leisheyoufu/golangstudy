package main

import (
	"log"
	"net/http"

	"github.com/leisheyoufu/golangstudy/rest/api"
)

func main() {
	nodeApi := api.NewNodeApi()
	log.Fatal(http.ListenAndServe(":8080", nodeApi.Router))
}
