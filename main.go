package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml2 "gopkg.in/yaml.v2"
)

const (
	OutputTypeDotenv OutputType = "dotenv"
)

var (
	DefaultManifestYamlFiles []string = []string{
		"./secrets.yaml",
		"./secrets.yml",
	}
)

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
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	ValueFrom   ValueFrom `yaml:"valueFrom"`
}

type ValueFrom struct {
	AwsParameterStore ValueFromAwsParameterStoreConfig `yaml:"awsParameterStore,omitempty"`
}

type ValueFromAwsParameterStoreConfig struct {
	Key string `yaml:"key"`
}

type OutputConfig struct {
	Type OutputType `yaml:"type"`
	Path string     `yaml:"path"`
}

type OutputType string

func main() {
	// read manifest file
	var file []byte
	for _, filename := range DefaultManifestYamlFiles {
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
		panic(fmt.Errorf("failed to find manifest file paths=%v", DefaultManifestYamlFiles))
	}

	// parse
	m := Manifest{}
	err := yaml2.Unmarshal(file, &m)
	if err != nil {
		panic(fmt.Errorf("failed to parse manifest yaml. %v", err))
	}

	// validate manifest config

	// debug print
	debugFormatJSON("manifest", m)
}

func debugFormatJSON(header string, i interface{}) {
	b, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		fmt.Printf("JSON Error: %s", err.Error())
		return
	}
	fmt.Printf("%s: %s", header, string(b))
}
