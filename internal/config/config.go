package config

import (
	"log/slog"
	"os"
)

type Config struct {
	AppEnv       string
	HTTPAddr     string
	DatabasePath string
	LogLevel     slog.Level
	WebOrigin    string
	WebStaticDir string
}

func Load() Config {
	port := getenv("PORT", "")
	httpAddr := getenv("RAHAT_HTTP_ADDR", ":8080")
	if port != "" {
		httpAddr = ":" + port
	}

	return Config{
		AppEnv:       getenv("APP_ENV", "development"),
		HTTPAddr:     httpAddr,
		DatabasePath: getenv("DATABASE_PATH", "./var/rahat.sqlite3"),
		LogLevel:     parseLogLevel(getenv("LOG_LEVEL", "info")),
		WebOrigin:    getenv("WEB_ORIGIN", "http://localhost:5200"),
		WebStaticDir: getenv("WEB_STATIC_DIR", ""),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parseLogLevel(value string) slog.Level {
	switch value {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
