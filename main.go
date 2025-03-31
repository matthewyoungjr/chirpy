package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is the response for the main dir...."))
	})

	server.ListenAndServe()
}
