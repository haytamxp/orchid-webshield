package models

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/haytamxp/orchid-webshield/webapp/database"
	_ "modernc.org/sqlite"
)

func newAdminTestDB(t *testing.T) *sql.DB {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	// Une seule connexion garantit que les tests utilisent
	// la même base SQLite en mémoire.
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec(database.Schema); err != nil {
		conn.Close()
		t.Fatalf("apply test schema: %v", err)
	}

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

func createAdminTestUser(t *testing.T, conn *sql.DB, username string) int64 {
	t.Helper()

	id, err := CreateUser(
		conn,
		username,
		username+"@example.com",
		"test-password-hash",
	)
	if err != nil {
		t.Fatalf("create user %q: %v", username, err)
	}

	return id
}

func TestPromoteFirstAdmin(t *testing.T) {
	conn := newAdminTestDB(t)

	createAdminTestUser(t, conn, "alice")
	createAdminTestUser(t, conn, "bob")

	if err := PromoteFirstAdmin(conn, "alice"); err != nil {
		t.Fatalf("promote first admin: %v", err)
	}

	alice, err := GetUserByUsername(conn, "alice")
	if err != nil {
		t.Fatalf("get alice: %v", err)
	}
	if alice.Role != RoleAdmin {
		t.Fatalf("alice role = %q, want %q", alice.Role, RoleAdmin)
	}

	err = PromoteFirstAdmin(conn, "bob")
	if !errors.Is(err, ErrAdminAlreadyExists) {
		t.Fatalf("second promotion error = %v, want ErrAdminAlreadyExists", err)
	}
}

func TestPromoteFirstAdminUnknownUser(t *testing.T) {
	conn := newAdminTestDB(t)

	err := PromoteFirstAdmin(conn, "unknown")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestUpdateUserRole(t *testing.T) {
	conn := newAdminTestDB(t)

	aliceID := createAdminTestUser(t, conn, "alice")
	bobID := createAdminTestUser(t, conn, "bob")

	if err := UpdateUserRole(conn, aliceID, RoleAdmin); err != nil {
		t.Fatalf("promote alice: %v", err)
	}

	// Alice est le seul ADMIN : sa rétrogradation doit être refusée.
	err := UpdateUserRole(conn, aliceID, RoleUser)
	if !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("demotion error = %v, want ErrLastAdmin", err)
	}

	// Bob devient ADMIN : Alice peut désormais redevenir USER.
	if err := UpdateUserRole(conn, bobID, RoleAdmin); err != nil {
		t.Fatalf("promote bob: %v", err)
	}

	if err := UpdateUserRole(conn, aliceID, RoleUser); err != nil {
		t.Fatalf("demote alice while bob remains admin: %v", err)
	}

	alice, err := GetUserByID(conn, aliceID)
	if err != nil {
		t.Fatalf("get alice: %v", err)
	}
	if alice.Role != RoleUser {
		t.Fatalf("alice role = %q, want %q", alice.Role, RoleUser)
	}
}

func TestUpdateUserRoleRejectsInvalidRoleAndUnknownUser(t *testing.T) {
	conn := newAdminTestDB(t)

	id := createAdminTestUser(t, conn, "alice")

	if err := UpdateUserRole(conn, id, "SUPERADMIN"); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("invalid role error = %v, want ErrInvalidRole", err)
	}

	if err := UpdateUserRole(conn, 9999, RoleAdmin); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown user error = %v, want ErrNotFound", err)
	}
}

func TestListUsersForAdmin(t *testing.T) {
	conn := newAdminTestDB(t)

	createAdminTestUser(t, conn, "alice")
	createAdminTestUser(t, conn, "bob")

	users, err := ListUsersForAdmin(conn)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("number of users = %d, want 2", len(users))
	}

	if users[0].Username != "alice" || users[1].Username != "bob" {
		t.Fatalf("unexpected user order: %+v", users)
	}

	if users[0].Role != RoleUser || users[1].Role != RoleUser {
		t.Fatalf("new users should have role USER: %+v", users)
	}
}