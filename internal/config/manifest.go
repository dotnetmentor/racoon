package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dotnetmentor/racoon/internal/output"
	yaml2 "gopkg.in/yaml.v2"
)

const (
	OutputTypeDotenv OutputType = "dotenv"
	OutputTypeTfvars OutputType = "tfvars"
	OutputTypeJson   OutputType = "json"

	StoreTypeAwsParameterStore string = "awsParameterStore"
)

func NewManifest(paths []string) (Manifest, error) {
	// read manifest file
	var file []byte
	for _, filename := range paths {
		mp, _ := filepath.Abs(filename)

		if _, err := os.Stat(mp); os.IsNotExist(err) {
			continue
		}

		bs, err := ioutil.ReadFile(mp)
		if err != nil {
			return Manifest{}, fmt.Errorf("failed to read manifest file (path=%s). %v", mp, err)
		}
		file = bs
	}

	if file == nil {
		return Manifest{}, fmt.Errorf("failed to find manifest file paths=%v", paths)
	}

	// parse
	m := Manifest{}
	err := yaml2.Unmarshal(file, &m)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest yaml. %v", err)
	}

	// TODO: validate manifest config

	return m, nil
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
	KmsKey           string `yaml:"kmsKey"`
	DefaultKeyFormat string `yaml:"keyFormat"`
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

type OutputType string

type OutputConfig struct {
	Type    OutputType             `yaml:"type,omitempty"`
	Path    string                 `yaml:"path"`
	Map     map[string]string      `yaml:"map"`
	Include []string               `yaml:"include"`
	Exclude []string               `yaml:"exclude"`
	Config  map[string]interface{} `yaml:"config"`
	output  output.Output
}

func (o *OutputConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type raw OutputConfig
	if err := unmarshal((*raw)(o)); err != nil {
		return err
	}

	output, err := UnmarshalConfig(o.Type, o.Config)
	if err != nil {
		return err
	}
	o.output = output

	return nil
}
