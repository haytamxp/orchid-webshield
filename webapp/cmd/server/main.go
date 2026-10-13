package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/haytamxp/orchid-webshield/webapp/internal/config"
	"github.com/haytamxp/orchid-webshield/webapp/internal/db"
	"github.com/haytamxp/orchid-webshield/webapp/internal/handlers"
	"github.com/haytamxp/orchid-webshield/webapp/internal/middleware"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer conn.Close()

	var tables int
	conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables)
	log.Printf("database ready: %d tables", tables)

		h := &handlers.Handler{DB: conn, CookieSecure: cfg.CookieSecure}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /register", h.Register)
    mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", middleware.RequireAuth(conn, h.Logout))
	mux.HandleFunc("GET /me", middleware.RequireAuth(conn, h.Me))
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("webapp listening on :%s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}