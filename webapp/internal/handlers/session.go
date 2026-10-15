
package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
)

// sessionCookieName est le nom du cookie de session.
const sessionCookieName = "session"

// dummyHash évite de révéler l'existence d'un compte par le temps de réponse.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// logSecurityEvent écrit un événement sans interrompre la requête en cas d'erreur.
func (h *Handler) logSecurityEvent(
	r *http.Request,
	eventType string,
	severity string,
	username *string,
	status int,
	details map[string]any,
) {
	if h.Events == nil {
		return
	}

	if err := h.Events.LogRequest(
		r, eventType, severity, username, status, details,
	); err != nil {
		log.Printf("security event logging failed: %v", err)
	}
}

// Login vérifie les identifiants, crée une session et pose le cookie.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))

	user, err := models.GetUserByUsername(h.DB, username)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		log.Printf("login: lookup error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	hash := dummyHash
	if user != nil {
		hash = []byte(user.PasswordHash)
	}
	passwordErr := bcrypt.CompareHashAndPassword(hash, []byte(req.Password))

	if user == nil || passwordErr != nil {
		h.logSecurityEvent(
			r,
			"login_failure",
			"MEDIUM",
			nil,
			http.StatusUnauthorized,
			map[string]any{"reason": "invalid_credentials"},
		)

		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
		return
	}

	token, expires, err := models.CreateSession(h.DB, user.ID)
	if err != nil {
		log.Printf("login: create session error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	eventUsername := user.Username
	h.logSecurityEvent(
		r,
		"login_success",
		"LOW",
		&eventUsername,
		http.StatusOK,
		nil,
	)

	writeData(w, http.StatusOK, map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}
