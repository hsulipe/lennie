package httphandler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/hsulipe/lennie/login/internal/services"
)

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// OIDCLogin redirects the browser to the Dex authorization endpoint.
func (h *Handler) OIDCLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
		Path:     "/",
	})
	http.Redirect(w, r, h.oidcConfig.AuthCodeURL(state), http.StatusFound)
}

// OIDCCallback handles the redirect from Dex, exchanges the code for an ID
// token, verifies it, and returns (or provisions) the matching user.
func (h *Handler) OIDCCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		writeError(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	token, err := h.oidcConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		writeError(w, "failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		writeError(w, "missing id_token in token response", http.StatusInternalServerError)
		return
	}

	idToken, err := h.oidcVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		writeError(w, "id_token verification failed", http.StatusUnauthorized)
		return
	}

	var claims struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		writeError(w, "failed to parse token claims", http.StatusInternalServerError)
		return
	}

	user, err := h.oidcService.SignIn(services.OIDCSignInInput{
		Subject:       idToken.Subject,
		Email:         claims.Email,
		Name:          claims.Name,
		EmailVerified: claims.EmailVerified,
	})
	if err != nil {
		if err.Error() == "email already registered with another provider" {
			writeError(w, err.Error(), http.StatusConflict)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, user, http.StatusOK)
}
