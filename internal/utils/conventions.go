package utils

import (
	"strings"

	"github.com/fatih/camelcase"
)

func CamelCaseSplitToUpperJoinByUnderscore(name string) (key string) {
	parts := camelcase.Split(name)
	for i, part := range parts {
		parts[i] = strings.ToUpper(part)
	}
	key = strings.Join(parts, "_")
	return
}

func CamelCaseSplitToLowerJoinByUnderscore(name string) (key string) {
	parts := camelcase.Split(name)
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
	key = strings.Join(parts, "_")
	return
}
