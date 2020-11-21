package config

import (
	"os"
	"strings"
)

// GetEnvVar looks up env var and returns its value or fallback if not available
func GetEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
