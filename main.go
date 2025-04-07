package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	serverHits atomic.Int32
}

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

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK %d", http.StatusOK)
	})

	mux.HandleFunc("/api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errmsg := Error{Error: "Method not allowed !"}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(errmsg)
			return
		}

		var requestBody RequestBody
		json.NewDecoder(r.Body).Decode(&requestBody)

		if len(requestBody.Body) > 140 {
			errmsg := &Error{Error: "Chirp is too long"}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&errmsg)
			return
		}

		response := Valid{Valid: true}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&response)

	})

	log.Fatal(server.ListenAndServe())

}

func (a *apiConfig) MetricsMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.serverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}
func (a *apiConfig) Reset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.serverHits.Swap(0)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Count has been reset successfully."))
	})

}
