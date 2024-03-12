package api

import (
	"errors"
	"fmt"
)

func NewConfigurationError(msg string) *ConfigurationError {
	return &ConfigurationError{
		msg: msg,
	}
}

func NewNotFoundError(err error, key string, source SourceType) *NotFoundError {
	msg := "value not found"
	if err != nil {
		msg += ", see inner error for more details"
	}
	return &NotFoundError{
		msg:    msg,
		key:    key,
		inner:  err,
		source: source,
	}
}

func NewValidationError(msg string, val Value) *ValidationError {
	return &ValidationError{
		msg: msg,
		val: val,
	}
}

func NewFormattingError(msg string) *FormattingError {
	return &FormattingError{
		msg: msg,
	}
}

func WrapFormattingErrors(errs []*FormattingError) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return &FormattingError{
		errors: errs,
	}
}

func IsConfigruationError(err error) bool {
	if err == nil {
		return false
	}
	var configurationError *ConfigurationError
	switch {
	case errors.As(err, &configurationError):
		return true
	default:
		return false
	}
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var notFound *NotFoundError
	switch {
	case errors.As(err, &notFound):
		return true
	default:
		return false
	}
}

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		return true
	default:
		return false
	}
}

func IsFormattingError(err error) bool {
	if err == nil {
		return false
	}
	var formatErr *FormattingError
	switch {
	case errors.As(err, &formatErr):
		return true
	default:
		return false
	}
}

type ConfigurationError struct {
	msg string
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("ConfigurationError, %s", e.msg)
}

type NotFoundError struct {
	msg    string
	inner  error
	key    string
	source SourceType
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("NotFoundError, %s (source=%s key=%s)", e.msg, e.source, e.key)
}

func (e *NotFoundError) InnerError() error {
	return e.inner
}

type ValidationError struct {
	val Value
	msg string
}

func (e *ValidationError) Error() string {
	if e.val != nil {
		return fmt.Sprintf("ValidationError, %s (source=%s value=%s)", e.msg, e.val.Source(), e.val)
	}
	return fmt.Sprintf("ValidationError, %s (value=<nil>)", e.msg)
}

type FormattingError struct {
	msg    string
	errors []*FormattingError
}

func (e *FormattingError) Error() string {
	if len(e.errors) > 0 {
		str := "FormattingError,\n"
		for i, err := range e.errors {
			str += fmt.Sprintf("  %d) %s\n", i, err.msg)
		}
		return str
	}
	return fmt.Sprintf("FormattingError, %s", e.msg)
}
