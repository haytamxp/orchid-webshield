package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("WEBAPP_PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	log.Printf("webapp listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}