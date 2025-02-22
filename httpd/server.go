package main

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	if err := http.ListenAndServe(":9000", http.HandlerFunc(helloHandler)); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, strings.NewReader("hello world!\n"))
}
