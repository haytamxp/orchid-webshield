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

// dummyHash sert à faire un calcul bcrypt même quand le compte n'existe pas,
// pour que le temps de réponse ne révèle pas quels comptes existent.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

	writeData(w, http.StatusOK, map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}