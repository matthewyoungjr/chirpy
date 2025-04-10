package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	config := &apiConfig{}

	mux.Handle("/app/", config.MetricsMiddleWare(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`
			<html>

				<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
				</body>

			</html>
	`, config.serverHits.Load())))
	})

	mux.Handle("POST /admin/reset", config.Reset())

	mux.HandleFunc("GET /api/healthz", Healthz)

	mux.HandleFunc("/api/validate_chirp", ValidateChirp)

	log.Fatal(server.ListenAndServe())

}
