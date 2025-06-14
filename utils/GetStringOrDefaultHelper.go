package utils

import "strings"

func GetStringOrDefault(val interface{}, fallback string) string {
	if str, ok := val.(string); ok && strings.TrimSpace(str) != "" {
		return str
	}
	return fallback
}
