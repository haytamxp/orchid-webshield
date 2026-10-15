package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name         string
		user         *models.User
		requiredRole string
		wantStatus   int
		wantNext     bool
	}{
		{
			name:         "unauthenticated user receives 401",
			user:         nil,
			requiredRole: "ADMIN",
			wantStatus:   http.StatusUnauthorized,
			wantNext:     false,
		},
		{
			name: "USER cannot access ADMIN route",
			user: &models.User{
				ID:   1,
				Role: "USER",
			},
			requiredRole: "ADMIN",
			wantStatus:   http.StatusForbidden,
			wantNext:     false,
		},
		{
			name: "USER can access USER route",
			user: &models.User{
				ID:   1,
				Role: "USER",
			},
			requiredRole: "USER",
			wantStatus:   http.StatusNoContent,
			wantNext:     true,
		},
		{
			name: "ADMIN can access ADMIN route",
			user: &models.User{
				ID:   2,
				Role: "ADMIN",
			},
			requiredRole: "ADMIN",
			wantStatus:   http.StatusNoContent,
			wantNext:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			next := http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			handler := RequireRole(tt.requiredRole, next)
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.user != nil {
				ctx := context.WithValue(
					req.Context(),
					userKey,
					tt.user,
				)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf(
					"status = %d, want %d",
					rec.Code,
					tt.wantStatus,
				)
			}

			if nextCalled != tt.wantNext {
				t.Errorf(
					"nextCalled = %t, want %t",
					nextCalled,
					tt.wantNext,
				)
			}
		})
	}
}