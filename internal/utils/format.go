package utils

import (
	"strings"
)

type Formatting struct {
	WordSeparator string
	PathSeparator string
	Uppercase     bool
	Lowercase     bool
	Prefix        string
}

func FormatKey(s string, f Formatting) string {
	prefix := len(f.Prefix) > 0
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
	if prefix {
		fs = f.Prefix + fs
	}
	return fs
}
