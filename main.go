package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/fatih/camelcase"
	"github.com/urfave/cli/v2"
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

	// run commands using manifest
	app := &cli.App{
		Name:  "racoon",
		Usage: "working with secrets is my thing",
		Commands: []*cli.Command{
			{
				Name:  "export",
				Usage: "exports all secrets according to manifest",
				Action: func(c *cli.Context) error {
					if awsRegion := stringEnvVar("AWS_REGION", ""); awsRegion == "" {
						return fmt.Errorf("required environment variable AWS_REGION has no value set")
					}

					awsConfig, err := config.LoadDefaultConfig(c.Context)
					if err != nil {
						return err
					}

					ssmClient := ssm.NewFromConfig(awsConfig)

					// read from param store
					values := map[string]string{}
					for _, s := range m.Secrets {
						if s.Default != nil {
							fmt.Println("reading", s.Name, "from", "default")
							values[s.Name] = *s.Default
						}

						if s.ValueFrom != nil {
							if s.ValueFrom.AwsParameterStore != nil {
								fmt.Println("reading", s.Name, "from", "awsParameterStore")
								out, err := ssmClient.GetParameter(c.Context, &ssm.GetParameterInput{
									Name:           &s.ValueFrom.AwsParameterStore.Key,
									WithDecryption: true,
								})
								if err != nil {
									return err
								}
								values[s.Name] = *out.Parameter.Value
							}
						}
					}

					// create outputs
					for _, o := range m.Outputs {
						switch o.Type {
						case OutputTypeDotenv:
							fmt.Println("exporting secrets in dotenv format")
							var b strings.Builder
							for _, s := range m.Secrets {
								parts := camelcase.Split(s.Name)
								for i, part := range parts {
									parts[i] = strings.ToUpper(part)
								}
								key := strings.Join(parts, "_")
								value := strings.TrimSuffix(values[s.Name], "\n")
								fmt.Fprintf(&b, "%s=\"%s\"\n", key, value)
							}
							ioutil.WriteFile(o.Path, []byte(b.String()), 0600)
							break
						default:
							panic(fmt.Errorf("unsupported output type %s", o.Type))
						}
					}

					return nil
				},
			},
			{
				Name:  "debug",
				Usage: "prints debug output",
				Action: func(c *cli.Context) error {
					debugFormatJSON("manifest", m)
					return nil
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func debugFormatJSON(header string, i interface{}) {
	b, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		fmt.Printf("JSON Error: %s\n", err.Error())
		return
	}
	fmt.Printf("%s: %s\n", header, string(b))
}

func stringEnvVar(key, defaultValue string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return defaultValue
}
