package config

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DBPath           string
	SessionTime      time.Duration
	AppEncryptionKey []byte
	LogLevel         slog.Level
	CleanupInterval  time.Duration
	LockoutThreshold int
	LockoutDuration  time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{}

	cfg.DBPath = os.Getenv("DB_PATH")
	if cfg.DBPath == "" {
		return nil, fmt.Errorf("DB_PATH is required")
	}

	sessionTimeStr := os.Getenv("SESSION_TIME")
	if sessionTimeStr == "" {
		return nil, fmt.Errorf("SESSION_TIME is required")
	}
	st, err := time.ParseDuration(sessionTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TIME format: %w", err)
	}
	cfg.SessionTime = st

	keyBase64 := os.Getenv("APP_ENCRYPTION_KEY")
	if keyBase64 == "" {
		return nil, fmt.Errorf("APP_ENCRYPTION_KEY is required")
	}
	keyBytes, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 in APP_ENCRYPTION_KEY: %w", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("APP_ENCRYPTION_KEY must decode to exactly 32 bytes (AES-256)")
	}
	cfg.AppEncryptionKey = keyBytes

	logLevelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevelStr == "" {
		logLevelStr = "info"
	}
	switch logLevelStr {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid LOG_LEVEL %q; must be debug, info, warn, or error", logLevelStr)
	}

	cleanupStr := os.Getenv("CLEANUP_INTERVAL")
	if cleanupStr == "" {
		cleanupStr = "1m"
	}
	ct, err := time.ParseDuration(cleanupStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CLEANUP_INTERVAL format: %w", err)
	}
	cfg.CleanupInterval = ct

	lockoutThresholdStr := os.Getenv("LOCKOUT_THRESHOLD")
	if lockoutThresholdStr == "" {
		lockoutThresholdStr = "5"
	}
	lt, err := strconv.Atoi(lockoutThresholdStr)
	if err != nil {
		return nil, fmt.Errorf("invalid LOCKOUT_THRESHOLD: %w", err)
	}
	cfg.LockoutThreshold = lt

	lockoutDurationStr := os.Getenv("LOCKOUT_DURATION")
	if lockoutDurationStr == "" {
		lockoutDurationStr = "15m"
	}
	ld, err := time.ParseDuration(lockoutDurationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid LOCKOUT_DURATION format: %w", err)
	}
	cfg.LockoutDuration = ld

	return cfg, nil
}
