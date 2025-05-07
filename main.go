package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/matthewyoungjr/chirpy/internal/database"
)

var config *apiConfig

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	dbQueries := database.New(db)

	config = &apiConfig{
		DB: dbQueries,
	}

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
	mux.HandleFunc("POST /api/users", CreateUser)
	mux.HandleFunc("GET /api/chirps", GetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", GetChirp)
	mux.HandleFunc("POST /api/chirps", CreateChirp)

	log.Fatal(server.ListenAndServe())

}
