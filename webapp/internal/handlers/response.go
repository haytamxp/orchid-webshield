package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON écrit une réponse JSON avec le code HTTP donné.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeData écrit une réponse de succès : {"data": ...}
func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, map[string]any{"data": data})
}

// writeError écrit une réponse d'erreur : {"error": {"code": ..., "message": ...}}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// decodeJSON lit le corps de la requête en JSON, avec une taille maximale de 1 Mo
// et en refusant les champs inconnus.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}