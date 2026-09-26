package main

import (
	"log"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/assets/logo.png", http.FileServer(http.Dir(".")))

	server := http.Server{Handler: serveMux, Addr: ":8080"}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
