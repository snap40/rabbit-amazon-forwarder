package utils

import "os"

func GetEnvOrDefault(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	return v
}
