package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
	"github.com/haytamxp/orchid-webshield/webapp/internal/validation"
)

// bcryptCost règle la lenteur du hachage : plus il est élevé, plus la force brute est coûteuse.
const bcryptCost = 12

// Handler regroupe les dépendances des routes.
type Handler struct {
	DB           *sql.DB
	CookieSecure bool
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register crée un compte avec le rôle USER.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	if err := validation.Username(username); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	email, err := validation.Email(req.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := validation.Password(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		log.Printf("register: hash error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	id, err := models.CreateUser(h.DB, username, email, string(hash))
	if errors.Is(err, models.ErrDuplicate) {
		writeError(w, http.StatusConflict, "conflict", "username or email already in use")
		return
	}
	if err != nil {
		log.Printf("register: create user error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	writeData(w, http.StatusCreated, map[string]any{
		"id":       id,
		"username": username,
		"email":    email,
		"role":     "USER",
	})
}