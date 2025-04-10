package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	serverHits atomic.Int32
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
