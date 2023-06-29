package config

import (
	"fmt"
	"strings"
)

func ReplaceParams(s string, ls Parameters) string {
	for k, v := range ls {
		s = strings.ReplaceAll(s, fmt.Sprintf("{%s}", k), v)
	}
	return s
}
