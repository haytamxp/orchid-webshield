package models

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	// ErrNotFound : aucun enregistrement ne correspond.
	ErrNotFound = errors.New("not found")
	// ErrDuplicate : le username ou l'email existe déjà.
	ErrDuplicate = errors.New("already exists")
)

// User représente une ligne de la table users.
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Role         string
}

// CreateUser insère un nouvel utilisateur avec le rôle USER.
// Le mot de passe doit déjà être haché : ce package ne voit jamais le mot de passe en clair.
func CreateUser(db *sql.DB, username, email, passwordHash string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		username, email, passwordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, ErrDuplicate
		}
		return 0, err
	}
	return result.LastInsertId()
}

// GetUserByUsername retrouve un utilisateur par son nom.
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	return scanUser(db.QueryRow(
		"SELECT id, username, email, password_hash, role FROM users WHERE username = ?",
		username,
	))
}

// GetUserByID retrouve un utilisateur par son identifiant.
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	return scanUser(db.QueryRow(
		"SELECT id, username, email, password_hash, role FROM users WHERE id = ?",
		id,
	))
}

func scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}