package config

import (
	"github.com/sirupsen/logrus"
)

var (
	DefaultManifestYamlFiles []string = []string{
		"./secrets.yaml",
		"./secrets.yml",
	}
)

type AppContext struct {
	Log      *logrus.Logger
	Manifest Manifest
}

func NewContext() AppContext {
	m := NewManifest(DefaultManifestYamlFiles)

	l := logrus.New()
	l.Formatter = &PrefixedTextFormatter{
		Prefix: "racoon ",
	}

	c := AppContext{
		Log:      l,
		Manifest: m,
	}

	return c
}
