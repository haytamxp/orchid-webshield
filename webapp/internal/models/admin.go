package models

import (
	"database/sql"
	"errors"
	"strings"
)

const (
	RoleUser  = "USER"
	RoleAdmin = "ADMIN"
)

var (
	ErrAdminAlreadyExists = errors.New("an admin already exists")
	ErrLastAdmin          = errors.New("cannot demote the last admin")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidUsername    = errors.New("username is required")
)

// AdminUser contient uniquement les informations publiques
// nécessaires à l'administration des comptes.
type AdminUser struct {
	ID       int64
	Username string
	Email    string
	Role     string
}

// ListUsersForAdmin retourne la liste des utilisateurs sans les mots de passe.
func ListUsersForAdmin(db *sql.DB) ([]AdminUser, error) {
	rows, err := db.Query(`
		SELECT id, username, email, role
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]AdminUser, 0)

	for rows.Next() {
		var user AdminUser

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// PromoteFirstAdmin promeut un utilisateur existant seulement
// si aucun administrateur n'existe encore.
func PromoteFirstAdmin(db *sql.DB, username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return ErrInvalidUsername
	}

	result, err := db.Exec(`
		UPDATE users
		SET role = ?
		WHERE username = ?
		  AND NOT EXISTS (
			SELECT 1 FROM users WHERE role = ?
		  )
	`, RoleAdmin, username, RoleAdmin)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 1 {
		return nil
	}

	var adminCount int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE role = ?`,
		RoleAdmin,
	).Scan(&adminCount); err != nil {
		return err
	}

	if adminCount > 0 {
		return ErrAdminAlreadyExists
	}

	return ErrNotFound
}

// UpdateUserRole modifie le rôle d'un compte et empêche
// la rétrogradation du dernier administrateur.
func UpdateUserRole(db *sql.DB, userID int64, role string) error {
	if role != RoleUser && role != RoleAdmin {
		return ErrInvalidRole
	}

	var result sql.Result
	var err error

	if role == RoleUser {
		// La condition SQL empêche de rétrograder le dernier ADMIN.
		result, err = db.Exec(`
			UPDATE users
			SET role = ?
			WHERE id = ?
			  AND (
				role <> ?
				OR (
					SELECT COUNT(*)
					FROM users
					WHERE role = ?
				) > 1
			  )
		`, RoleUser, userID, RoleAdmin, RoleAdmin)
	} else {
		result, err = db.Exec(`
			UPDATE users
			SET role = ?
			WHERE id = ?
		`, RoleAdmin, userID)
	}

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected > 0 {
		return nil
	}

	// Si aucune ligne n'a été modifiée, distinguons les erreurs.
	var currentRole string
	err = db.QueryRow(
		`SELECT role FROM users WHERE id = ?`,
		userID,
	).Scan(&currentRole)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Une demande qui ne change rien est considérée comme valide.
	if currentRole == role {
		return nil
	}

	if currentRole == RoleAdmin && role == RoleUser {
		return ErrLastAdmin
	}

	return errors.New("role update did not affect a row")
}
