package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app/", config.MetricsMiddleWare(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /admin/metrics", GetMetrics)

	mux.Handle("POST /admin/reset", config.Reset())

	mux.HandleFunc("GET /api/healthz", Healthz)

	mux.HandleFunc("/api/validate_chirp", ValidateChirp)

	log.Fatal(server.ListenAndServe())

}
