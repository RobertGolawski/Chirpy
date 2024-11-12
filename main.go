package main

import (
	"net/http"
)

func main() {
	var server = http.NewServeMux()
	var serverStruct = http.Server{
		Handler: server,
		Addr:    ":8080",
	}
	serverStruct.ListenAndServe()

}
