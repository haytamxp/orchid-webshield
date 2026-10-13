package handlers

import (
	"log"
	"net/http"

	"github.com/haytamxp/orchid-webshield/webapp/internal/middleware"
	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
)

// Me renvoie l'utilisateur authentifié. La route doit être protégée par RequireAuth.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	writeData(w, http.StatusOK, map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// Logout supprime la session côté serveur et efface le cookie.
// La route doit être protégée par RequireAuth.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err == nil {
		if err := models.DeleteSession(h.DB, cookie.Value); err != nil {
			log.Printf("logout: delete session error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	writeData(w, http.StatusOK, map[string]string{"status": "logged out"})
}