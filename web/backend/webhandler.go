package backend

import (
	"log"
	"net/http"
	"os"
)

type WebHandler struct {
}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

func (handler *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Receive %s request from %s%s", r.Method, r.Host, r.URL.Path)
	if r.URL.EscapedPath() == "/" || r.URL.EscapedPath() == "/index.html" {
		// Do not store the html page in the cache. If the user is to click on 'switch language',
		// we want a different index.html (for the right locale) to be served when the page refreshes.
		w.Header().Add("Cache-Control", "no-store")
	}
	acceptLanguage := os.Getenv("ACCEPT_LANGUAGE")
	if acceptLanguage == "" {
		acceptLanguage = r.Header.Get("Accept-Language")
	}
	http.FileServer(http.Dir("public")).ServeHTTP(w, r)
}
