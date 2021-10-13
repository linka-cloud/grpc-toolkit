package service

import (
	"net/http"

	"github.com/justinas/alice"
)

type ServeMux interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

type Middleware = alice.Constructor
