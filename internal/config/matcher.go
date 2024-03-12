package config

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MatchTypeNotSet    MatchType = ""
	MatchEqual         MatchType = " = "
	MatchNotEqual      MatchType = " != "
	MatchRegexEqual    MatchType = " =~ "
	MatchRegexNotEqual MatchType = " !~ "
)

var (
	operators []MatchType = []MatchType{
		MatchRegexNotEqual,
		MatchRegexEqual,
		MatchNotEqual,
		MatchEqual,
	}
)

type MatchType string

type Matcher struct {
	operator   MatchType
	expression string
	regex      *regexp.Regexp
}

func (m Matcher) Match(s string) bool {
	switch m.operator {
	case MatchEqual:
		return m.expression == s
	case MatchNotEqual:
		return m.expression != s
	case MatchRegexEqual:
		return m.regex.MatchString(s)
	case MatchRegexNotEqual:
		return !m.regex.MatchString(s)
	}
	return false
}

func ParseExpression(expr string) (key string, matcher Matcher, err error) {
	for _, op := range operators {
		p := strings.SplitN(expr, string(op), 2)
		if len(p) == 2 {
			key := strings.TrimSpace(p[0])
			v := strings.TrimSpace(p[1])
			if v == `""` {
				v = ""
			}
			m := Matcher{
				operator:   op,
				expression: v,
			}
			switch m.operator {
			case MatchRegexEqual, MatchRegexNotEqual:
				r, err := regexp.Compile(m.expression)
				if err != nil {
					return key, Matcher{}, err
				}
				m.regex = r
			}
			return key, m, nil
		}
	}
	return "", Matcher{}, fmt.Errorf("invalid expression, %s", expr)
}
