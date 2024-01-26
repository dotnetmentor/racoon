package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotnetmentor/racoon/internal/utils"
)

type parameters map[string]string

type OrderedParameterList []Parameter

type Parameter struct {
	Key   string
	Value string
}

func ParseParams(ls []string) (parameters, error) {
	p := parameters{}
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

func (p parameters) ValidateParams(pl ParameterConfigList) error {
	// validate parameters are defined by manifest
	for k := range p {
		if ok := pl.HasKey(k); !ok {
			return fmt.Errorf("parameter %s, provided but not defined", k)
		}
	}

	reserved := []string{"name", "key"}

	// validate parameters
	for _, pc := range pl {
		pv, ok := p[pc.Key]

		if utils.StringSliceContains(reserved, pc.Key) {
			return fmt.Errorf("parameter key \"%s\" is reserved and cannot be used", pc.Key)
		}

		if pc.Required {
			if !ok {
				return fmt.Errorf("required parameter must be set, parameter: %s", pc.Key)
			}
		}

		if ok {
			if len(pc.Regexp) > 0 {
				re, err := regexp.Compile(pc.Regexp)
				if err != nil {
					return fmt.Errorf("invalid regular expression for parameter %s, err: %w", pc.Key, err)
				}

				if !re.MatchString(pv) {
					return fmt.Errorf("parameter %s, regular expression validation failed (value=%s regexp=%s)", pc.Key, pv, pc.Regexp)
				}
			}
		}
	}

	return nil
}

func (p parameters) Ordered(pl ParameterConfigList) (ordered OrderedParameterList) {
	for _, pc := range pl {
		if v, ok := p[pc.Key]; ok {
			ordered = append(ordered, Parameter{
				Key:   pc.Key,
				Value: v,
			})
		}
	}
	return ordered
}

func (op OrderedParameterList) replace(s string) string {
	for _, p := range op {
		s = strings.ReplaceAll(s, fmt.Sprintf("{%s}", p.Key), p.Value)
	}
	return s
}

func (op OrderedParameterList) Value(key string) (string, bool) {
	for _, p := range op {
		if p.Key == key {
			return p.Value, true
		}
	}
	return "", false
}
