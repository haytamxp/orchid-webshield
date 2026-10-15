package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/haytamxp/orchid-webshield/webapp/database"
	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
		"github.com/haytamxp/orchid-webshield/webapp/internal/middleware"
	_ "modernc.org/sqlite"
)

// newHandlerTestDB crée une base SQLite en mémoire avec le vrai schéma.
func newHandlerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	conn.SetMaxOpenConns(1) // une seule connexion = une seule base en mémoire

	if _, err := conn.Exec(database.Schema); err != nil {
		conn.Close()
		t.Fatalf("apply test schema: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return conn
}

// createHandlerTestUser crée un utilisateur (rôle USER par défaut).
func createHandlerTestUser(t *testing.T, conn *sql.DB, username string) int64 {
	t.Helper()

	id, err := models.CreateUser(conn, username, username+"@example.com", "test-password-hash")
	if err != nil {
		t.Fatalf("create user %q: %v", username, err)
	}
	return id
}

// newAdminTestMux enregistre les handlers admin comme le fait main.go,
// mais sans les middlewares : ces tests couvrent la logique du handler.
func newAdminTestMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/users", h.ListUsers)
	mux.HandleFunc("PATCH /admin/users/{id}/role", h.UpdateUserRole)
	return mux
}

// doRequest envoie une requête au mux et renvoie la réponse enregistrée.
func doRequest(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// errorCode extrait le champ error.code d'une réponse d'erreur JSON.
func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", rec.Body.String(), err)
	}
	return body.Error.Code
}

func TestListUsersHidesSecrets(t *testing.T) {
	conn := newHandlerTestDB(t)
	createHandlerTestUser(t, conn, "alice")

	h := &Handler{DB: conn}
	rec := doRequest(newAdminTestMux(h), http.MethodGet, "/admin/users", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"username":"alice"`) {
		t.Errorf("response should contain alice: %s", body)
	}
	if !strings.Contains(body, `"role":"USER"`) {
		t.Errorf("response should contain role USER: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "password") ||
		strings.Contains(body, "test-password-hash") {
		t.Errorf("response must not expose password data: %s", body)
	}
}

func TestUpdateUserRoleSuccess(t *testing.T) {
	conn := newHandlerTestDB(t)
	createHandlerTestUser(t, conn, "alice")
	bobID := createHandlerTestUser(t, conn, "bob")

	// alice devient le premier admin, pour que bob puisse être modifié
	// sans toucher au dernier administrateur.
	if err := models.PromoteFirstAdmin(conn, "alice"); err != nil {
		t.Fatalf("promote first admin: %v", err)
	}

	h := &Handler{DB: conn}
	path := "/admin/users/" + strconv.FormatInt(bobID, 10) + "/role"
	rec := doRequest(newAdminTestMux(h), http.MethodPatch, path, `{"role":"ADMIN"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var role string
	if err := conn.QueryRow(`SELECT role FROM users WHERE id = ?`, bobID).Scan(&role); err != nil {
		t.Fatalf("read role: %v", err)
	}
	if role != "ADMIN" {
		t.Errorf("role in database = %q, want ADMIN", role)
	}
}
func TestUpdateUserRoleErrors(t *testing.T) {
	conn := newHandlerTestDB(t)
	aliceID := createHandlerTestUser(t, conn, "alice")
	createHandlerTestUser(t, conn, "bob")

	// alice est l'unique administrateur.
	if err := models.PromoteFirstAdmin(conn, "alice"); err != nil {
		t.Fatalf("promote first admin: %v", err)
	}

	h := &Handler{DB: conn}
	mux := newAdminTestMux(h)
	alicePath := "/admin/users/" + strconv.FormatInt(aliceID, 10) + "/role"

	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{"invalid role", "/admin/users/2/role", `{"role":"SUPERADMIN"}`, http.StatusBadRequest, "invalid_role"},
		{"unknown user", "/admin/users/999/role", `{"role":"ADMIN"}`, http.StatusNotFound, "not_found"},
		{"non numeric id", "/admin/users/abc/role", `{"role":"ADMIN"}`, http.StatusBadRequest, "invalid_user_id"},
		{"negative id", "/admin/users/-1/role", `{"role":"ADMIN"}`, http.StatusBadRequest, "invalid_user_id"},
		{"unknown json field", "/admin/users/2/role", `{"role":"ADMIN","is_admin":true}`, http.StatusBadRequest, "invalid_request"},
		{"malformed json", "/admin/users/2/role", `{"role":`, http.StatusBadRequest, "invalid_request"},
		{"last admin", alicePath, `{"role":"USER"}`, http.StatusConflict, "last_admin"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(mux, http.MethodPatch, tc.path, tc.body)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if got := errorCode(t, rec); got != tc.wantCode {
				t.Errorf("error code = %q, want %q", got, tc.wantCode)
			}
		})
	}

	// Aucun de ces refus ne doit avoir modifié la base : alice reste ADMIN.
	var role string
	if err := conn.QueryRow(`SELECT role FROM users WHERE id = ?`, aliceID).Scan(&role); err != nil {
		t.Fatalf("read role: %v", err)
	}
	if role != "ADMIN" {
		t.Errorf("alice role = %q, want ADMIN", role)
	}
}
// newProtectedAdminMux monte la vraie chaîne de main.go :
// RequireAuth -> RequireRole(ADMIN) -> handler.
func newProtectedAdminMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/users",
		middleware.RequireAuth(h.DB,
			middleware.RequireRole(models.RoleAdmin, h.ListUsers)))
	mux.HandleFunc("PATCH /admin/users/{id}/role",
		middleware.RequireAuth(h.DB,
			middleware.RequireRole(models.RoleAdmin, h.UpdateUserRole)))
	return mux
}

// doRequestWithSession envoie une requête avec un cookie de session (ou sans, si token est vide).
func doRequestWithSession(mux http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: token})
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAdminRoutesAuthorization(t *testing.T) {
	conn := newHandlerTestDB(t)
	aliceID := createHandlerTestUser(t, conn, "alice") // restera USER
	createHandlerTestUser(t, conn, "bob")              // deviendra ADMIN

	if err := models.PromoteFirstAdmin(conn, "bob"); err != nil {
		t.Fatalf("promote first admin: %v", err)
	}

	aliceToken, _, err := models.CreateSession(conn, aliceID)
	if err != nil {
		t.Fatalf("create alice session: %v", err)
	}
	var bobID int64
	if err := conn.QueryRow(`SELECT id FROM users WHERE username = 'bob'`).Scan(&bobID); err != nil {
		t.Fatalf("read bob id: %v", err)
	}
	bobToken, _, err := models.CreateSession(conn, bobID)
	if err != nil {
		t.Fatalf("create bob session: %v", err)
	}

	mux := newProtectedAdminMux(&Handler{DB: conn})
	selfPath := "/admin/users/" + strconv.FormatInt(aliceID, 10) + "/role"

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		token      string
		wantStatus int
	}{
		{"anonymous list", http.MethodGet, "/admin/users", "", "", http.StatusUnauthorized},
		{"anonymous patch", http.MethodPatch, selfPath, `{"role":"ADMIN"}`, "", http.StatusUnauthorized},
		{"user list", http.MethodGet, "/admin/users", "", aliceToken, http.StatusForbidden},
		{"user self-promotion", http.MethodPatch, selfPath, `{"role":"ADMIN"}`, aliceToken, http.StatusForbidden},
		{"admin list", http.MethodGet, "/admin/users", "", bobToken, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequestWithSession(mux, tc.method, tc.path, tc.body, tc.token)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}

	// L'auto-promotion refusée ne doit pas avoir modifié la base.
	var role string
	if err := conn.QueryRow(`SELECT role FROM users WHERE id = ?`, aliceID).Scan(&role); err != nil {
		t.Fatalf("read role: %v", err)
	}
	if role != "USER" {
		t.Errorf("alice role = %q, want USER", role)
	}
}