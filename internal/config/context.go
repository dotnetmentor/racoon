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
	Metadata   AppMetadata
	Manifest   Manifest
	Parameters Parameters
}

type AppMetadata struct {
	Version string
	Commit  string
	Date    string
}

func NewContext(metadata AppMetadata, paths ...string) (AppContext, error) {
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
		Metadata: metadata,
		Manifest: m,
	}

	return c, err
}
