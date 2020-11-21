package config

import (
	"os"
	"strconv"
	"strings"
)

// GetEnvVar looks up env var and returns its value or fallback if not available
func GetEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}

// GetEnvBoolVar returns value from env var based on the key or falls back to provided value
func GetEnvBoolVar(key string, fallbackValue bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	b, err := strconv.ParseBool(val)
	if err == nil {
		return b
	}
	return fallbackValue
}
