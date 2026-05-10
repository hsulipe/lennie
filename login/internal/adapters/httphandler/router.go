package httphandler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/health", h.HealthCheck).Methods(http.MethodGet)
	r.HandleFunc("/signin", h.SignIn).Methods(http.MethodPost)
	r.HandleFunc("/signup", h.SignUp).Methods(http.MethodPost)
	return r
}
