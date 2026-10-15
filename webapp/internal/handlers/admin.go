package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
)

// adminUserResponse définit les champs publics renvoyés par l'API.
type adminUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// UpdateRoleRequest contient le nouveau rôle demandé.
type updateRoleRequest struct {
	Role string `json:"role"`
}

// ListUsers retourne la liste des utilisateurs.
// La route devra être réservée aux ADMIN.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := models.ListUsersForAdmin(h.DB)
	if err != nil {
		log.Printf("admin: list users error: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"internal error",
		)
		return
	}

	response := make([]adminUserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, adminUserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		})
	}

	writeData(w, http.StatusOK, response)
}

// UpdateUserRole modifie le rôle d'un utilisateur existant.
// La route devra être réservée aux ADMIN.
func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_user_id",
			"invalid user ID",
		)
		return
	}

	var req updateRoleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"invalid request body",
		)
		return
	}

	err = models.UpdateUserRole(h.DB, id, req.Role)
	switch {
	case errors.Is(err, models.ErrInvalidRole):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_role",
			"role must be USER or ADMIN",
		)
		return

	case errors.Is(err, models.ErrNotFound):
		writeError(
			w,
			http.StatusNotFound,
			"not_found",
			"user not found",
		)
		return

	case errors.Is(err, models.ErrLastAdmin):
		writeError(
			w,
			http.StatusConflict,
			"last_admin",
			"cannot demote the last administrator",
		)
		return

	case err != nil:
		log.Printf("admin: update user role error: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"internal error",
		)
		return
	}

	writeData(w, http.StatusOK, map[string]any{
		"user_id": id,
		"role":    req.Role,
	})
}
