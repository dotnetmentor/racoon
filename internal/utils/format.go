package utils

import (
	"strings"
)

type Formatting struct {
	WordSeparator string
	PathSeparator string
	Uppercase     bool
	Lowercase     bool
}

func FormatKey(s string, f Formatting) string {
	fs := ""
	parts := SplitPath(s)
	for i, path := range parts {
		words := SplitCamelCase(path)
		for i, word := range words {
			if f.Uppercase {
				word = strings.ToUpper(word)
			} else if f.Lowercase {
				word = strings.ToLower(word)
			}

			words[i] = word
		}
		fs += strings.Join(words, f.WordSeparator)

		if i+1 < len(parts) {
			fs += f.PathSeparator
		}
	}
	return fs
}
