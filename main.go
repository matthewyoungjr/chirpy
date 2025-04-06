package main

import (
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
