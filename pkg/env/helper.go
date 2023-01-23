package env

import (
	"os"
	"strconv"
	"strings"
)

func EnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func IntEnvOrDefault(key string, def int) int {
	varEnv := os.Getenv(key)
	val, err := strconv.Atoi(varEnv)
	if err != nil {
		return def
	}
	return val
}

func BoolEnvOrDefault(key string, def bool) bool {
	varEnv := os.Getenv(key)
	if len(varEnv) == 0 {
		return def
	}
	if strings.ToLower(varEnv) == "true" || varEnv == "1" {
		return true
	}
	return false
}
