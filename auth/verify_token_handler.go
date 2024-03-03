package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}

	var validationResponse = struct {
		Email string `json:"email"`
		Role  string `json:"role"`
		Valid bool   `json:"valid"`
	}{}

	ok, validatedToken, err := jwtToken.Validate(tokenString)
	if !ok {
		log.Printf(`verify, invalid token=%s: err=%s`, tokenString, err)
	} else {
		validationResponse.Email = validatedToken.Email
		validationResponse.Role = validatedToken.Role
		validationResponse.Valid = true
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(validationResponse); err != nil {
		http.Error(w, "Failed to encode validation response as JSON", http.StatusInternalServerError)
		log.Printf(`failed to encode validation response (%+v): %s`, validationResponse, err)
		return
	}
}
