package config

import (
	"bufio"
	"os"
	"strings"
)

// Config contient tous les réglages de l'application.
type Config struct {
	Port          string
	DatabasePath  string
	UploadDir     string
	LogFile       string
	SessionSecret string
	CookieSecure  bool
}

// Load lit le fichier .env (s'il existe), puis les variables d'environnement.
func Load() Config {
	loadEnvFile()

	return Config{
		Port:          getEnv("WEBAPP_PORT", "8081"),
		DatabasePath:  getEnv("WEBAPP_DATABASE_PATH", "database/webshield.db"),
		UploadDir:     getEnv("WEBAPP_UPLOAD_DIR", "uploads"),
		LogFile:       getEnv("WEBAPP_LOG_FILE", "logs/security_events.jsonl"),
		SessionSecret: getEnv("WEBAPP_SESSION_SECRET", "change-me"),
		CookieSecure:  getEnv("WEBAPP_COOKIE_SECURE", "false") == "true",
	}
}

// getEnv renvoie la variable d'environnement, ou la valeur par défaut si elle est vide.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// loadEnvFile lit un fichier de lignes CLE=valeur et les place dans l'environnement.
// Une variable déjà définie n'est pas écrasée.
const envFile = ".env"

func loadEnvFile() {
	file, err := os.Open(envFile)
	if err != nil {
		return // pas de fichier .env : on utilise les valeurs par défaut
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		if os.Getenv(key) == "" {
			os.Setenv(key, strings.TrimSpace(value))
		}
	}
}