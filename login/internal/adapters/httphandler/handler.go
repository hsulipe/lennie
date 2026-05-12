package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/hsulipe/lennie/login/internal/services"
	"golang.org/x/oauth2"
)

type Handler struct {
	signIn       *services.SignInService
	signUp       *services.SignUpService
	oidcService  *services.OIDCService
	oidcConfig   *oauth2.Config
	oidcVerifier *oidc.IDTokenVerifier
	dbConnected  bool
}

func NewHandler(
	signIn *services.SignInService,
	signUp *services.SignUpService,
	oidcService *services.OIDCService,
	oidcConfig *oauth2.Config,
	oidcVerifier *oidc.IDTokenVerifier,
	dbConnected bool,
) *Handler {
	return &Handler{
		signIn:       signIn,
		signUp:       signUp,
		oidcService:  oidcService,
		oidcConfig:   oidcConfig,
		oidcVerifier: oidcVerifier,
		dbConnected:  dbConnected,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"healthy": true, "db_connected": h.dbConnected}, http.StatusOK)
}

type signInRequest struct {
	Provider    string `json:"provider"`
	ProviderID  string `json:"provider_id"`
	Credentials string `json:"credentials"`
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := h.signIn.SignIn(req.Provider, req.ProviderID, req.Credentials)
	if err != nil {
		if err.Error() == "invalid credentials" {
			writeError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, user, http.StatusOK)
}

type signUpRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	CPF         string `json:"cpf"`
	Phone       string `json:"phone"`
	Birthdate   string `json:"birthdate"`
	Provider    string `json:"provider"`
	ProviderID  string `json:"provider_id"`
	Credentials string `json:"credentials"`
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req signUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := h.signUp.SignUp(services.SignUpInput{
		Name:       req.Name,
		Email:      req.Email,
		CPF:        req.CPF,
		Phone:      req.Phone,
		Birthdate:  req.Birthdate,
		Provider:   req.Provider,
		ProviderID: req.ProviderID,
		Password:   req.Credentials,
	})
	if err != nil {
		switch err.Error() {
		case "email already registered":
			writeError(w, err.Error(), http.StatusConflict)
		case "invalid birthdate format, expected YYYY-MM-DD":
			writeError(w, err.Error(), http.StatusBadRequest)
		default:
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, user, http.StatusCreated)
}

func writeJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error": %q}`, msg)
}
