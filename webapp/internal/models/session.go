package models

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"time"
)

// SessionDuration est la durée de vie d'une session.
const SessionDuration = 24 * time.Hour

// hashToken calcule l'empreinte SHA-256 d'un jeton. La base ne stocke que cette empreinte.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateSession crée une session pour l'utilisateur et renvoie le jeton (à mettre dans le cookie)
// ainsi que sa date d'expiration. Le jeton lui-même n'est jamais stocké.
func CreateSession(db *sql.DB, userID int64) (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := time.Now().UTC().Add(SessionDuration)

	_, err := db.Exec(
		"INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)",
		userID, hashToken(token), expires.Format(time.RFC3339),
	)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

// GetUserBySession renvoie l'utilisateur lié au jeton, si la session existe et n'est pas expirée.
func GetUserBySession(db *sql.DB, token string) (*User, error) {
	return scanUser(db.QueryRow(
		`SELECT u.id, u.username, u.email, u.password_hash, u.role
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = ? AND s.expires_at > ?`,
		hashToken(token), time.Now().UTC().Format(time.RFC3339),
	))
}

// DeleteSession supprime la session correspondant au jeton.
func DeleteSession(db *sql.DB, token string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE token_hash = ?", hashToken(token))
	return err
}