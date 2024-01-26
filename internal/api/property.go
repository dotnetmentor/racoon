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

	exists := false
	for _, ep := range properties {
		if ep.Name == property.Name {
			exists = true

			// Allowing new property to be sensitive while enforcing it if existing property is marked sensitive
			if !property.sensitive {
				property.sensitive = ep.sensitive
			}

			// Copy from existing property
			if len(property.Description) > 0 && property.Description != ep.Description {
				apiLog.Warnf("%s/%s, overriding description is not allowed, description already defined in %s", property.source, property.Name, ep.source)
			}
			property.Description = ep.Description

			if property.rules != config.DefaultPropertyRules && property.rules != ep.rules {
				apiLog.Warnf("%s/%s, overriding rules is not allowed, rules already defined in %s", property.source, property.Name, ep.source)
			}
			property.rules = ep.rules

			break
		}
	}

	if !exists {
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
		return NewValidationError(fmt.Sprintf("value must not be nil for property %s", p.Name), v)
	}

	if IsNotFoundError(v.Error()) {
		return NewValidationError(fmt.Sprintf("value not found for property %s", p.Name), v)
	}

	if v.Error() != nil {
		return NewValidationError(fmt.Sprintf("value resolved with error for property %s, %v", v.Error(), p.Name), v)
	}

	if len(v.Raw()) == 0 && !p.Rules().Validation.AllowEmpty {
		return NewValidationError(fmt.Sprintf("empty value not allowed for property %s", p.Name), v)
	}

	return nil
}
