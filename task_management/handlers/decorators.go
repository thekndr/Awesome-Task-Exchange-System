package handlers

import (
	"github.com/thekndr/ates/auth_client"
	"github.com/thekndr/ates/common"
	"log"
	"net/http"
)

func WithUserId(handler HandlerWithUserId) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.URL.Query().Get("token")

		userId, err := auth_client.GetClaimWithoutVerification[string](tokenString, "id")
		if err != nil || userId == "" {
			http.Error(w, "failed to identify user", http.StatusInternalServerError)
			log.Printf(`user-id is missing in token: %s`, err)
			return
		}

		handler.Handle(userId, w, r)
	}
}

func WithManagerRoleOnly(handler RegularHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.URL.Query().Get("token")

		role, err := auth_client.GetClaimWithoutVerification[string](tokenString, "role")
		if err != nil || role == "" {
			http.Error(w, "failed to identify role", http.StatusInternalServerError)
			log.Printf(`role is missing in token: %s`, err)
			return
		}

		if common.Role(role) != common.ManagerRole {
			http.Error(w, "Invalid role for the operation", http.StatusForbidden)
			log.Printf(`not a manager: %s`, role)
			return
		}

		handler.Handle(w, r)
	}
}
