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

type User struct {
	Email string `json:"email"`
}

type CreateChirpParam struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}
