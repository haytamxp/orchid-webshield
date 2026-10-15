
package middleware

import (
	"log"
	"net/http"

	"github.com/haytamxp/orchid-webshield/webapp/internal/securityevents"
)

// RequireRole conserve la compatibilité avec les routes existantes.
func RequireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return RequireRoleWithEvents(requiredRole, nil, next)
}

// RequireRoleWithEvents contrôle le rôle et journalise les refus d'autorisation.
func RequireRoleWithEvents(
	requiredRole string,
	events *securityevents.Logger,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())

		if !ok || user == nil {
			logAuthorizationDenied(events, r, nil, http.StatusUnauthorized, "authentication_required")
			writeError(
				w,
				http.StatusUnauthorized,
				"unauthorized",
				"authentication required",
			)
			return
		}

		if user.Role != requiredRole {
			username := user.Username
			logAuthorizationDenied(events, r, &username, http.StatusForbidden, "insufficient_permissions")
			writeError(
				w,
				http.StatusForbidden,
				"forbidden",
				"insufficient permissions",
			)
			return
		}

		next(w, r)
	}
}

func logAuthorizationDenied(
	events *securityevents.Logger,
	r *http.Request,
	username *string,
	status int,
	reason string,
) {
	if events == nil {
		return
	}

	if err := events.LogRequest(
		r,
		"authorization_denied",
		"MEDIUM",
		username,
		status,
		map[string]any{"reason": reason},
	); err != nil {
		log.Printf("security event logging failed: %v", err)
	}
}
