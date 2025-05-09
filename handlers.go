package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/matthewyoungjr/chirpy/internal/auth"
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
	var u UserParams

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		log.Println("Error decoding body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !strings.Contains(u.Email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(u.Password)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	params := database.CreateUserParams{
		Email:          u.Email,
		HashedPassword: hash,
	}

	log.Println("Creating user with email:", u.Email)
	user, err := config.DB.CreateUser(r.Context(), params)

	if err != nil {
		log.Printf("Error : %v", err)
		http.Error(w, "Could not create user", http.StatusBadRequest)
		return
	}

	response := struct {
		ID    uuid.UUID
		Email string
	}{
		ID:    user.ID,
		Email: user.Email,
	}

	log.Println("User created:", user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&response)

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

func GetChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("id")
	chirpID, err := uuid.Parse(chirpId)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	chirp, err := config.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Chirp not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chirp)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var u UserParams

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)
	if err != nil {
		log.Printf("Couldn't decode json : %v", err)
		return
	}

	if !strings.Contains(u.Email, "@") {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	user, err := config.DB.GetUserByEmail(r.Context(), u.Email)
	if err != nil {
		log.Printf("User lookup failed: %v", err)
		http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	if err = auth.CheckPasswordHash(user.HashedPassword, u.Password); err != nil {
		log.Printf("Password mismatch : %v", err)
		http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	response := struct {
		ID    uuid.UUID `json:"id"`
		Email string    `json:"email"`
	}{
		ID:    user.ID,
		Email: user.Email,
	}

	log.Println("User logged in:", user.Email)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}
