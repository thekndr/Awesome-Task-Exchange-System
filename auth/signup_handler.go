package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Signup struct {
	EventCh chan interface{}
}

func (h *Signup) Handle(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the role to "worker" by default if it's not specified
	if user.Role == "" {
		user.Role = "worker" // Set the default role to "worker"
	} else {
		if !ensureValidRole(user.Role, w) {
			return
		}
	}

	var publicId string
	err := db.QueryRow("INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING uuid", user.Email, user.Password, user.Role).Scan(&publicId)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		log.Printf("insertion failed (email=%s,password=%d,role=%s): %s", user.Email, len(user.Password), user.Role, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User created successfully")

	go func() {
		h.EventCh <- Event{
			Name: "user-registered",
			Payload: map[string]interface{}{
				"role":    user.Role,
				"user-id": publicId,
			},
		}
	}()
}
