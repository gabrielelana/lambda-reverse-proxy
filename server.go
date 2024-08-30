package main

import "net/http"

func NewServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("pong"))
	})

	return mux
}
