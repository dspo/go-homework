package main

import (
	"log"
	"net/http"
)

const (
	address = ":8080"
)

func main() {
	log.Printf("The HTTP Server is going to run on %s\n", address)
	http.HandleFunc("/healthz", func(_ http.ResponseWriter, _ *http.Request) {})
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("failed to ListenAndServe on %s\n", address)
	}
}
