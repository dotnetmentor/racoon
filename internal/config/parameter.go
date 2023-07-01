package config

import (
	"fmt"
	"regexp"
	"strings"
)

type Parameters map[string]string

func ParseParams(ls []string) (Parameters, error) {
	p := Parameters{}
	for _, l := range ls {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			return p, fmt.Errorf("invalid parameter format %s, value must conform to <key>=<value>, parts: %v", l, parts)
		}
		lk := parts[0]
		lv := parts[1]
		if len(lk) < 1 {
			return p, fmt.Errorf("invalid parameter %s, key must not be empty", l)
		}
		p[lk] = lv
	}
	return p, nil
}

func (p Parameters) ValidateParams(c ParameterConfig) error {
	// validate parameters are defined by manifest
	for k := range p {
		if _, ok := c[k]; !ok {
			return fmt.Errorf("parameter %s, provided but not defined", k)
		}
	}

	// validate parameters
	for k, v := range c {
		pv, ok := p[k]

		if v.Required {
			if !ok {
				return fmt.Errorf("required parameter must be set, parameter: %s", k)
			}
		}

		if ok {
			if len(v.Regexp) > 0 {
				re, err := regexp.Compile(v.Regexp)
				if err != nil {
					return fmt.Errorf("invalid regular expression for parameter %s, err: %w", k, err)
				}

				if !re.MatchString(pv) {
					return fmt.Errorf("parameter %s, regular expression validation failed (value=%s regexp=%s)", k, pv, v.Regexp)
				}
			}
		}
	}

	return nil
}

func (p Parameters) Replace(s string) string {
	for k, v := range p {
		s = strings.ReplaceAll(s, fmt.Sprintf("{%s}", k), v)
	}
	return s
}
