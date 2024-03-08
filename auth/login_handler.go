package main

import (
	"encoding/json"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var publicId, dbPassword, role string
	err := db.QueryRow("SELECT public_id, password, role FROM users WHERE email = $1", user.Email).Scan(&publicId, &dbPassword, &role)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// We should hash and compare the password in a real application
	if user.Password != dbPassword {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := jwtToken.Issue(publicId, user.Email, role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	var response = struct {
		Token string `json:"token"`
	}{Token: token}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode token as JSON", http.StatusInternalServerError)
		return
	}
}
