package utils

import "os"

func StringEnvVar(key, defaultValue string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return defaultValue
}

func StringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
