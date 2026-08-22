package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(l LoginHandler) http.Handler {
	router := mux.NewRouter()
	router.Path("/login").Methods(http.MethodPost).HandlerFunc(l.Login)
	return router
}
