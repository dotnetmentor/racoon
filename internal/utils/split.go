package utils

import (
	"strings"

	"github.com/fatih/camelcase"
)

func SplitPath(s string) (parts []string) {
	return strings.Split(s, ".")
}

func SplitCamelCase(s string) (parts []string) {
	return camelcase.Split(s)
}
