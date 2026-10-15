package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/haytamxp/orchid-webshield/webapp/internal/config"
	"github.com/haytamxp/orchid-webshield/webapp/internal/db"
	"github.com/haytamxp/orchid-webshield/webapp/internal/models"
)

func main() {
	os.Exit(run())
}

// run contient toute la logique et renvoie le code de sortie du programme.
func run() int {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/promote-admin <username>")
		return 2
	}
	username := os.Args[1]

	cfg := config.Load()
	conn, err := db.Open(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		return 1
	}
	defer conn.Close()

	err = models.PromoteFirstAdmin(conn, username)
	switch {
	case err == nil:
		fmt.Printf("OK: %q est maintenant ADMIN\n", username)
		return 0
	case errors.Is(err, models.ErrAdminAlreadyExists):
		fmt.Fprintln(os.Stderr, "refuse: un administrateur existe deja")
		return 1
	case errors.Is(err, models.ErrNotFound):
		fmt.Fprintf(os.Stderr, "refuse: utilisateur %q introuvable\n", username)
		return 1
	case errors.Is(err, models.ErrInvalidUsername):
		fmt.Fprintln(os.Stderr, "refuse: nom d'utilisateur invalide")
		return 2
	default:
		fmt.Fprintf(os.Stderr, "erreur: %v\n", err)
		return 1
	}
}