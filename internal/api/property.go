package api

import (
	"fmt"

	"github.com/dotnetmentor/racoon/internal/config"
)

func NewProperty(properties PropertyList, name, description, source string, sensitive bool, rules config.RuleConfig, formatting []config.FormattingConfig) (property Property, isNew bool) {
	property = Property{
		Name:        name,
		Description: description,
		source:      source,
		sensitive:   sensitive,
		rules:       rules,
		formatting:  formatting,
		values:      make(ValueList, 0),
	}

	found := false
	for _, p := range properties {
		if p.Name == property.Name {
			found = true
			if !property.sensitive {
				property.sensitive = p.sensitive
			}
			property.rules = p.rules
			break
		}
	}

	if !found {
		isNew = true
	}
	return
}

type PropertyList []Property

type Property struct {
	Name        string
	Description string

	source     string
	values     ValueList
	sensitive  bool
	rules      config.RuleConfig
	formatting []config.FormattingConfig
}

func (p *Property) Value() Value {
	max := len(p.values)
	for i := max - 1; i >= 0; i-- {
		v := p.values[i]
		// Exclude ValueNotFoundError's
		if v.Error() != nil && IsNotFoundError(v.Error()) {
			continue
		}
		return v
	}

	// If no value at this point, return first possible value
	if max > 0 {
		return p.values[max-1]
	}

	// No values
	return nil
}

func (p *Property) Values() ValueList {
	return p.values
}

func (p *Property) SetValue(val Value) Value {
	if val != nil {
		p.values = append(p.values, val)
	}
	return val
}

func (p *Property) SetRawValue(layer Layer, source SourceType, key, val string, err error, forceSensitive bool) Value {
	return p.SetValue(NewValue(NewValueSource(layer, source), key, val, err, p.sensitive || forceSensitive))
}

func (p *Property) String() string {
	return fmt.Sprintf("%s/%s", p.source, p.Name)
}

func (p Property) Source() string {
	return p.source
}

func (p Property) Sensitive() bool {
	return p.sensitive
}

func (p Property) Rules() config.RuleConfig {
	return p.rules
}

func (p Property) Formatting() []config.FormattingConfig {
	return p.formatting
}

func (p Property) WritableFormatters() (writable []config.FormattingConfig) {
	for _, fc := range p.Formatting() {
		if SourceType(fc.Source.SourceType()).Writable() {
			writable = append(writable, fc)
		}
	}
	return
}

func (p Property) Validate(v Value) error {
	if v == nil {
		return NewValidationError("value must not be nil", v)
	}

	if IsNotFoundError(v.Error()) {
		return NewValidationError("value not found", v)
	}

	if v.Error() != nil {
		return NewValidationError(fmt.Sprintf("value resolved with error, %v", v.Error()), v)
	}

	if len(v.Raw()) == 0 && !p.Rules().Validation.AllowEmpty {
		return NewValidationError(fmt.Sprintf("empty value not allowed for property %s", p.Name), v)
	}
	return nil
}
