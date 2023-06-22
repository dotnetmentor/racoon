package config

import (
	"context"

	"github.com/sirupsen/logrus"
)

var (
	DefaultManifestYamlFiles []string = []string{
		"./racoon.yaml",
		"./racoon.yml",
	}
)

type AppContext struct {
	Context    context.Context
	Log        *logrus.Logger
	Manifest   Manifest
	Parameters Parameters
}

type Parameters map[string]string

func NewContext(paths ...string) (AppContext, error) {
	l := logrus.New()
	l.Formatter = &PrefixedTextFormatter{
		Prefix: "racoon ",
	}

	if len(paths) == 0 {
		return AppContext{
			Log: l,
		}, nil
	}

	m, err := NewManifest(paths)
	c := AppContext{
		Log:      l,
		Manifest: m,
	}

	return c, err
}
