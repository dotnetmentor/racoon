package config

import "testing"

var cases = []struct {
	name       string
	input      string
	key        string
	operator   MatchType
	expression string
	regex      bool
	error      bool
}{
	{"invalid operator whitespace", "key=value", "", MatchTypeNotSet, "", false, true},
	{"valid equal whitespace", "key = value", "key", MatchEqual, "value", false, false},
	{"valid not equal", "key != value", "key", MatchNotEqual, "value", false, false},
	{"valid regex equal", "key =~ (foo|bar)", "key", MatchRegexEqual, "(foo|bar)", true, false},
	{"valid regex not equal", "key !~ (foo|bar)", "key", MatchRegexNotEqual, "(foo|bar)", true, false},
	{"invalid expression string", "str", "", MatchTypeNotSet, "", false, true},
	{"invalid expression operator 1", "key ~~ (foo|bar)", "", MatchTypeNotSet, "", false, true},
	{"invalid expression operator 2", "key =! value", "", MatchTypeNotSet, "", false, true},
	{"valid expression without value", `key != ""`, "key", MatchNotEqual, "", false, false},
}

func TestParseExpression(t *testing.T) {
	for _, c := range cases {
		key, matcher, err := ParseExpression(c.input)
		if err != nil && !c.error {
			t.Errorf("(case: %s input:%s), unexpected error, %v", c.name, c.input, err)
			continue
		} else if err != nil && c.error {
			continue
		}

		if key != c.key {
			t.Errorf("case: %s, input=%s, unexpected key %s, expected=%s", c.name, c.input, key, c.key)
			continue
		}
		if matcher.operator != c.operator {
			t.Errorf("case: %s, input=%s, unexpected operator %s, expected=%s", c.name, c.input, matcher.operator, c.operator)
			continue
		}
		if matcher.expression != c.expression {
			t.Errorf("case: %s, input=%s, unexpected key %s, expected=%s", c.name, c.input, matcher.expression, c.expression)
			continue
		}
		if c.regex && matcher.regex == nil {
			t.Errorf("case: %s, input=%s, expected compiled regexp  for %s", c.name, c.input, matcher.expression)
			continue
		}
	}
}
