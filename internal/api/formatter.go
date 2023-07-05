package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/sirupsen/logrus"
)

func NewFormatter(f config.FormattingConfig, log *logrus.Logger) ValueFormatter {
	if f.Replace != nil {
		return &ReplaceFormatter{
			baseFormatter{
				log:    log,
				key:    *f.Replace,
				source: f.Source,
			},
		}
	}
	if f.RegexpReplace != nil {
		return &RegexpReplaceFormatter{
			baseFormatter{
				log:    log,
				key:    *f.RegexpReplace,
				source: f.Source,
			},
		}
	}
	return nil
}

type ValueFormatter interface {
	FormattingKey() string
	Source() *config.ValueSourceConfig
	Apply(format string, val Value) (str string, err error)
	String() string
}

type baseFormatter struct {
	log    *logrus.Logger
	key    string
	source *config.ValueSourceConfig
}

func (f baseFormatter) FormattingKey() string {
	return f.key
}

func (f baseFormatter) Source() *config.ValueSourceConfig {
	return f.source
}

type ReplaceFormatter struct {
	baseFormatter
}

func (f *ReplaceFormatter) String() string {
	return fmt.Sprintf("%T:\"%s\"", *f, f.FormattingKey())
}

func (f *ReplaceFormatter) Apply(format string, val Value) (string, error) {
	rkey := fmt.Sprintf("{%s}", f.key)
	f.log.Debugf("replacing %s with value: %s", rkey, val.String())
	return strings.ReplaceAll(format, rkey, val.Raw()), nil
}

type RegexpReplaceFormatter struct {
	baseFormatter
}

func (f *RegexpReplaceFormatter) String() string {
	return fmt.Sprintf("%T:\"%s\"", *f, f.FormattingKey())
}

func (f *RegexpReplaceFormatter) Apply(format string, val Value) (string, error) {
	r, err := regexp.Compile(f.key)
	if err != nil {
		return format, err
	}

	f.log.Debugf("regexp replacing %s with value: %s", f.key, val.String())
	return r.ReplaceAllLiteralString(format, val.Raw()), nil
}
