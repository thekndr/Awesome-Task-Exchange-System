package handlers

import (
	"net/http"
)

type HandlerWithUserId interface {
	Handle(userId string, w http.ResponseWriter, r *http.Request)
}

type RegularHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}
