package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml2 "gopkg.in/yaml.v2"
)

const (
	OutputTypeDotenv OutputType = "dotenv"
	OutputTypeTfvars OutputType = "tfvars"
)

func NewManifest(paths []string) Manifest {
	// read manifest file
	var file []byte
	for _, filename := range paths {
		mp, _ := filepath.Abs(filename)

		if _, err := os.Stat(mp); os.IsNotExist(err) {
			continue
		}

		bs, err := ioutil.ReadFile(mp)
		if err != nil {
			panic(fmt.Errorf("failed to read manifest file (path=%s). %v", mp, err))
		}
		file = bs
	}

	if file == nil {
		panic(fmt.Errorf("failed to find manifest file paths=%v", paths))
	}

	// parse
	m := Manifest{}
	err := yaml2.Unmarshal(file, &m)
	if err != nil {
		panic(fmt.Errorf("failed to parse manifest yaml. %v", err))
	}

	// TODO: validate manifest config

	return m
}

type Manifest struct {
	Stores  StoresConfig   `yaml:"stores"`
	Secrets []SecretConfig `yaml:"secrets"`
	Outputs []OutputConfig `yaml:"outputs"`
}

type StoresConfig struct {
	AwsParameterStore AwsParameterStoreConfig `yaml:"awsParameterStore,omitempty"`
}

type AwsParameterStoreConfig struct {
	KmsKey string `yaml:"kmsKey"`
}

type SecretConfig struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Default     *string    `yaml:"default,omitempty"`
	ValueFrom   *ValueFrom `yaml:"valueFrom,omitempty"`
}

type ValueFrom struct {
	AwsParameterStore *ValueFromAwsParameterStoreConfig `yaml:"awsParameterStore,omitempty"`
}

type ValueFromAwsParameterStoreConfig struct {
	Key string `yaml:"key"`
}

type OutputConfig struct {
	Type OutputType `yaml:"type"`
	Path string     `yaml:"path"`
}

type OutputType string
