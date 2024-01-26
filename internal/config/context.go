package config

import (
	"context"

	"github.com/dotnetmentor/racoon/internal/environment"
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
	Metadata   AppMetadata
	Manifest   Manifest
	Parameters OrderedParameterList
}

type AppMetadata struct {
	Version string
	Commit  string
	Date    string
}

func NewContext(metadata AppMetadata, paths ...string) (AppContext, error) {
	l := logrus.New()
	l.Formatter = &PrefixedTextFormatter{
		Prefix: "[racoon] ",
	}

	if len(paths) == 0 {
		return AppContext{
			Log: l,
		}, nil
	}

	m, err := NewManifest(paths)
	c := AppContext{
		Log:      l,
		Metadata: metadata,
		Manifest: m,
	}
	if err != nil {
		return c, err
	}

	backendEnabled, err := environment.BoolVar("RACOON_BACKEND_ENABLED", m.Backend.Enabled)
	if err != nil {
		return c, err
	}

	if backendEnabled != m.Backend.Enabled {
		if backendEnabled {
			c.Log.Warn("enabled backend, RACOON_BACKEND_ENABLED environment variable set to true")
		} else {
			c.Log.Warn("disabled backend, RACOON_BACKEND_ENABLED environment variable set to false")
		}
	}

	return c, nil
}
