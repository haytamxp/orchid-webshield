package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
	"github.com/haytamxp/orchid-webshield/webapp/internal/securityevents"
)

// SessionCookieName est le nom du cookie de session.
const SessionCookieName = "session"

type contextKey string

const userKey contextKey = "user"

// UserFromContext renvoie l'utilisateur authentifié placé par RequireAuth.
func UserFromContext(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(userKey).(*models.User)
	return u, ok
}

// RequireAuth conserve la vérification de session et journalise les refus.
func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuthWithEvents(db, nil, next)
}

// RequireAuthWithEvents vérifie la session et journalise les refus d'authentification.
func RequireAuthWithEvents(
	db *sql.DB,
	events *securityevents.Logger,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			logAuthorizationDenied(events, r, nil, http.StatusUnauthorized, "authentication_required")
			writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}

		user, err := models.GetUserBySession(db, cookie.Value)
		if errors.Is(err, models.ErrNotFound) {
			logAuthorizationDenied(events, r, nil, http.StatusUnauthorized, "invalid_session")
			writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}
		if err != nil {
			log.Printf("auth: session lookup error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}

		ctx := context.WithValue(r.Context(), userKey, user)
		next(w, r.WithContext(ctx))
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}