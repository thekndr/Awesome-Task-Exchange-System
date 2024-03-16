package main

import (
	"encoding/json"
	"fmt"
	"github.com/thekndr/ates/event_streaming"
	"log"
	"net/http"
)

type Signup struct {
	EventCh chan event_streaming.InternalEvent
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
	err := db.QueryRow("INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING public_id", user.Email, user.Password, user.Role).Scan(&publicId)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		log.Printf("insertion failed (email=%s,password=%d,role=%s): %s", user.Email, len(user.Password), user.Role, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User created successfully")

	go func() {
		h.EventCh <- event_streaming.InternalEvent{
			Name: "user-registered",
			Context: map[string]interface{}{
				"role":    user.Role,
				"user-id": publicId,
				"email":   user.Email,
			},
		}
	}()
}
