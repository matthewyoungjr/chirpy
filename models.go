package main

type RequestBody struct {
	Body string `json:"body"`
}

type Error struct {
	Error string `json:"error"`
}

type Valid struct {
	Valid bool `json:"valid"`
}

type Cleaned struct {
	CleanedBody string `json:"cleaned_body"`
}

type UserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateChirpParam struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}
