package main

import (
	"log"
	"net/http"
)

const (
	address = ":8080"
)

func main() {
	if err := http.ListenAndServe(address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})); err != nil {
		log.Fatalf("failed to ListenAndServe on %s\n", address)
	}
}
