package api

import (
	"fmt"
	"strings"

	"github.com/dotnetmentor/racoon/internal/backend"
	"github.com/dotnetmentor/racoon/internal/config"
)

type EncryptedConfig struct {
	backend    backend.Backend
	parameters config.OrderedParameterList

	Name       string              `json:"name" yaml:"name"`
	Labels     map[string]string   `json:"labels" yaml:"labels"`
	Properties []EncryptedProperty `json:"properties" yaml:"properties"`
}

type EncryptedProperty struct {
	Name        string  `json:"name" yaml:"name"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Sensitive   bool    `json:"sensitive,omitempty" yaml:"sensitive,omitempty"`
	Value       *string `json:"value" yaml:"value"`
}

func NewEncryptedConfig(m config.Manifest, p config.OrderedParameterList, backend backend.Backend) *EncryptedConfig {
	ec := EncryptedConfig{
		backend:    backend,
		parameters: p,

		Name:       m.Name,
		Labels:     make(map[string]string),
		Properties: make([]EncryptedProperty, 0),
	}
	for k, v := range m.Labels {
		ec.Labels[k] = p.Replace(v)
	}
	return &ec
}

func (ec *EncryptedConfig) Track(p Property) error {
	if ec.backend == nil {
		return nil
	}

	ep := EncryptedProperty{
		Name:        p.Name,
		Description: p.Description,
	}

	val := p.Value()

	if val == nil {
		ep.Value = nil
	} else {
		ep.Sensitive = val.Sensitive()

		// If validation passes but we have a not found error for the resolved value, skip export
		if !IsNotFoundError(val.Error()) {
			if val.Sensitive() && len(val.Raw()) > 0 {
				v, err := ec.backend.Encryption().Encrypt([]byte(val.Raw()))
				if err != nil {
					return err
				}
				ev := string(v)
				ep.Value = &ev
			} else {
				dv := val.Raw()
				ep.Value = &dv
			}
		}
	}

	ec.Properties = append(ec.Properties, ep)
	return nil
}

func (ec *EncryptedConfig) Path() string {
	path := make([]string, 0)
	for _, p := range ec.parameters {
		path = append(path, fmt.Sprintf("%s/%s", p.Key, p.Value))
	}
	path = append(path, fmt.Sprintf("%s/%s", ec.Name, "racoon.config"))
	return strings.Join(path, "/")
}
