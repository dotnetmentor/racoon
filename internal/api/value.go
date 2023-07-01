package api

import (
	"fmt"

	"github.com/ttacon/chalk"
)

func NewValue(source ValueSource, key string, val string, err error, sensitive bool) Value {
	if sensitive {
		return &SensitiveValue{
			source: source,
			key:    key,
			raw:    val,
			err:    err,
		}
	} else {
		return &ClearTextValue{
			source: source,
			key:    key,
			raw:    val,
			err:    err,
		}
	}
}

type ValueList []Value

func (l ValueList) Writable() (writable ValueList) {
	for _, v := range l {
		if len(v.Key()) > 0 && v.Source().Writable() {
			writable = append(writable, v)
		}
	}
	return
}

type Value interface {
	Sensitive() bool
	Source() ValueSource
	Key() string
	Raw() string
	Error() error
	SourceAndKey() string
	String() string
}

type SensitiveValue struct {
	source ValueSource
	key    string
	raw    string
	err    error
}

func (v *SensitiveValue) Sensitive() bool {
	return true
}

func (v *SensitiveValue) Source() ValueSource {
	return v.source
}

func (v *SensitiveValue) Key() string {
	return v.key
}

func (v *SensitiveValue) SourceAndKey() string {
	return fmt.Sprintf("%s (key=%s)", v.Source().String(), v.Key())
}

func (v *SensitiveValue) Raw() string {
	return v.raw
}

func (v *SensitiveValue) Error() error {
	return v.err
}

func (v *SensitiveValue) String() string {
	if v.err != nil && IsNotFoundError(v.err) {
		return fmt.Sprintf("%s%s%s", chalk.Yellow, "<not found>", chalk.ResetColor)
	}
	if v.err != nil {
		return fmt.Sprintf("%s%s%s", chalk.Red, "<error>", chalk.ResetColor)
	}
	return "<sensitive>"
}

type ClearTextValue struct {
	source ValueSource
	key    string
	raw    string
	err    error
}

func (v *ClearTextValue) Sensitive() bool {
	return false
}

func (v *ClearTextValue) Source() ValueSource {
	return v.source
}

func (v *ClearTextValue) Key() string {
	return v.key
}

func (v *ClearTextValue) SourceAndKey() string {
	return fmt.Sprintf("%s (key=%s)", v.Source().String(), v.Key())
}

func (v *ClearTextValue) Raw() string {
	return v.raw
}

func (v *ClearTextValue) Error() error {
	return v.err
}

func (v *ClearTextValue) String() string {
	if v.err != nil && IsNotFoundError(v.err) {
		return fmt.Sprintf("%s%s%s", chalk.Yellow, "<not found>", chalk.ResetColor)
	}
	if v.err != nil {
		return fmt.Sprintf("%s%s%s", chalk.Red, "<error>", chalk.ResetColor)
	}
	if len(v.raw) == 0 {
		return "<empty>"
	}
	return v.raw
}
