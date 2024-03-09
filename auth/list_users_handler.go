package main

import (
	"encoding/json"
	_ "github.com/thekndr/ates/auth_client"
	"log"
	"net/http"
	"time"
)

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	type UserInfo struct {
		PublicId  string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
	}

	// Query the database for all users
	rows, err := db.Query("SELECT public_id, created_at, email, role FROM users")
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		log.Printf(`query failed: %s`, err)
		return
	}
	defer rows.Close()

	var users []UserInfo

	for rows.Next() {
		var user UserInfo
		if err := rows.Scan(&user.PublicId, &user.CreatedAt, &user.Email, &user.Role); err != nil {
			http.Error(w, "Failed to load user data", http.StatusInternalServerError)
			log.Printf(`scan failed: %s`, err)
			return
		}
		users = append(users, user)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		http.Error(w, "Failed to read user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode users as JSON", http.StatusInternalServerError)
		return
	}
}
