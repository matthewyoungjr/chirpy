package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/matthewyoungjr/chirpy/internal/database"
)

type apiConfig struct {
	serverHits atomic.Int32
	DB         *database.Queries
}

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
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

	profane := []string{"kerfuffle", "sharbert", "fornax"}
	cleaned := requestBody.Body
	wasCleaned := false

	for _, p := range profane {
		if strings.Contains(cleaned, p) {
			cleaned = strings.ReplaceAll(cleaned, p, "****")
			wasCleaned = true
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if wasCleaned {
		newBody := Cleaned{CleanedBody: cleaned}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(newBody)
		return
	}

	response := Valid{Valid: true}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&response)

}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK %d", http.StatusOK)
}

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
		<html>

			<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
			</body>

		</html>
`, config.serverHits.Load())))
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

func CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")
	var email User

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&email); err != nil {
		log.Println("Error decoding body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !strings.Contains(email.Email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	log.Println("Creating user with email:", email.Email)
	user, err := config.DB.CreateUser(r.Context(), email.Email)

	if err != nil {
		log.Printf("Error : %v", err)
		http.Error(w, "Could not create user", http.StatusBadRequest)
		return
	}

	log.Println("User created:", user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&user)

}

func CreateChirp(w http.ResponseWriter, r *http.Request) {
	var request CreateChirpParam

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(request.Body) > 140 {
		http.Error(w, "Chirp is too long", http.StatusBadRequest)
		return
	}

	profane := []string{"kerfuffle", "sharbert", "fornax"}
	cleaned := request.Body

	for _, bad := range profane {
		cleaned = strings.ReplaceAll(cleaned, bad, "****")
	}

	chirp, err := config.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: uuid.MustParse(request.UserID),
	})

	if err != nil {
		log.Printf("Err : %v", err)
		http.Error(w, "could not create chirp", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chirp)
}

func GetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := config.DB.GetChirps(r.Context())
	if err != nil {
		http.Error(w, "Could not retrieve chirps", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chirps)
}
