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

func ParseParams(ls []string) (Parameters, error) {
	p := Parameters{}
	for _, l := range ls {
		parts := strings.Split(l, "=")
		// TODO: Error handling for flag value parsing
		lk := parts[0]
		lv := parts[1]
		p[lk] = lv
	}
	return p, nil
}
