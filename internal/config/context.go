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

func NewContext() (AppContext, error) {
	l := logrus.New()
	l.Formatter = &PrefixedTextFormatter{
		Prefix: "racoon ",
	}

	m, err := NewManifest(DefaultManifestYamlFiles)
	c := AppContext{
		Log:      l,
		Manifest: m,
	}

	return c, err
}
