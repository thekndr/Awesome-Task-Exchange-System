package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ChangeRole struct {
}

func (h *ChangeRole) Handle(w http.ResponseWriter, r *http.Request) {
	type changeRoleRequest struct {
		Email   string `json:"email"`
		NewRole string `json:"newRole"`
	}

	var requestData changeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !ensureValidRole(requestData.NewRole, w) {
		return
	}

	var publicId string
	err := db.QueryRow("UPDATE users SET role = $1 WHERE email = $2 RETURNING uuid", requestData.NewRole, requestData.Email).Scan(&publicId)
	if err != nil {
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User role updated successfully")
}