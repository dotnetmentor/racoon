package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/dotnetmentor/racoon/internal/output"
	"github.com/dotnetmentor/racoon/internal/utils"
	yaml2 "gopkg.in/yaml.v2"
)

const (
	SourceTypeAwsParameterStore SourceType = "awsParameterStore"
	SourceTypeEnvironment       SourceType = "env"
	SourceTypeLiteral           SourceType = "literal"
	SourceTypeParameter         SourceType = "parameter"

	OutputTypeDotenv OutputType = "dotenv"
	OutputTypeTfvars OutputType = "tfvars"
	OutputTypeJson   OutputType = "json"

	ExportTypeAll       ExportType = "all"
	ExportTypeSensitive ExportType = "sensitive"
	ExportTypeClearText ExportType = "cleartext"
)

type SourceType string

type OutputType string

type ExportType string

func NewManifest(paths []string) (Manifest, error) {
	// read manifest file
	var file []byte
	for _, filename := range paths {
		mp, _ := filepath.Abs(filename)

		if _, err := os.Stat(mp); os.IsNotExist(err) {
			continue
		}

		bs, err := os.ReadFile(mp)
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
	err := yaml2.UnmarshalStrict(file, &m)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest yaml. %v", err)
	}

	// TODO: Validate manifest config

	return m, nil
}

type Manifest struct {
	Config     Config         `yaml:"config"`
	Layers     []LayerConfig  `yaml:"layers"`
	Properties PropertyList   `yaml:"properties"`
	Outputs    []OutputConfig `yaml:"outputs"`
}

type Config struct {
	Parameters ParameterConfig `yaml:"parameters"`
	Sources    SourceConfig    `yaml:"sources"`
}

type LayerConfig struct {
	Name            string            `yaml:"name"`
	Match           map[string]string `yaml:"match"`
	Config          SourceConfig      `yaml:"config"`
	ImplicitSources []SourceType      `yaml:"implicitSources"`
	Properties      PropertyList      `yaml:"properties"`
}

type PropertyList []PropertyConfig

func (l PropertyList) Filter(excludes, includes []string) (properties PropertyList) {
	for _, p := range l {
		if len(excludes) > 0 && utils.StringSliceContains(excludes, p.Name) {
			continue
		}
		if len(includes) > 0 && !utils.StringSliceContains(includes, p.Name) {
			continue
		}
		properties = append(properties, p)
	}
	return
}

func (l PropertyList) Merge(pl PropertyList) (properties PropertyList) {
	properties = append(properties, l...)

	for _, p := range pl {
		if !utils.SliceContains(l, func(i PropertyConfig) bool {
			return i.Name == p.Name
		}) {
			properties = append(properties, p)
		}
	}
	return
}

func (l PropertyList) Remove(pl PropertyList) (properties PropertyList) {
	properties = append(properties, l...)

	return utils.SliceDelete(properties, func(i PropertyConfig) bool {
		return utils.SliceContains(pl, func(j PropertyConfig) bool {
			return i.Name == j.Name
		})
	})
}

type ParameterConfig map[string]ParameterRule

type ParameterRule struct {
	Required bool   `yaml:"required"`
	Regexp   string `yaml:"regexp"`
}

type SourceConfig struct {
	AwsParameterStore AwsParameterStoreConfig `yaml:"awsParameterStore"`
	Env               EnvConfig               `yaml:"env"`
}

type AwsParameterStoreConfig struct {
	ForceSensitive bool   `yaml:"forceSensitive"`
	KmsKey         string `yaml:"kmsKey"`
	DefaultKey     string `yaml:"defaultKey"`
}

func (c AwsParameterStoreConfig) Merge(config AwsParameterStoreConfig) AwsParameterStoreConfig {
	nc := AwsParameterStoreConfig{
		ForceSensitive: c.ForceSensitive,
		KmsKey:         c.KmsKey,
		DefaultKey:     c.DefaultKey,
	}

	if config.ForceSensitive {
		nc.ForceSensitive = true
	}

	if len(config.KmsKey) > 0 && nc.KmsKey != config.KmsKey {
		nc.KmsKey = config.KmsKey
	}

	if len(config.DefaultKey) > 0 && nc.DefaultKey != config.DefaultKey {
		nc.DefaultKey = config.DefaultKey
	}

	return nc
}

type EnvConfig struct {
	Dotfiles []string `yaml:"dotfiles"`
}

type PropertyConfig struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Default     *string            `yaml:"default,omitempty"`
	Sensitive   bool               `yaml:"sensitive,omitempty"`
	Source      *PropertyValueFrom `yaml:"source,omitempty"`
	Format      *FormattingConfig  `yaml:"format,omitempty"`
	Rules       RuleConfig         `yaml:"rules"`
}

func (s *PropertyConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawConfig PropertyConfig

	// Put defaults here
	raw := rawConfig{
		Rules: RuleConfig{
			Override: OverrideConfig{
				AllowImplicit: true,
				AllowExplicit: true,
			},
		},
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*s = PropertyConfig(raw)
	return nil
}

type FormattingConfig struct {
	Replace map[string]*PropertyValueFrom `yaml:"replace"`
}

type RuleConfig struct {
	Validation ValidationConfig `yaml:"validation"`
	Override   OverrideConfig   `yaml:"override"`
}

type ValidationConfig struct {
	AllowEmpty bool `yaml:"allowEmpty"`
}

type OverrideConfig struct {
	AllowImplicit bool `yaml:"allowImplicit"`
	AllowExplicit bool `yaml:"allowExplicit"`
}

type PropertyValueFrom struct {
	Parameter         *string                     `yaml:"parameter,omitempty"`
	Literal           *string                     `yaml:"literal,omitempty"`
	Environment       *ValueFromEvnironment       `yaml:"env,omitempty"`
	AwsParameterStore *ValueFromAwsParameterStore `yaml:"awsParameterStore,omitempty"`
}

type ValueFromEvnironment struct {
	Key string `yaml:"key"`
}

type ValueFromAwsParameterStore struct {
	Key string `yaml:"key"`
}

type OutputConfig struct {
	Type    OutputType             `yaml:"type"`
	Alias   string                 `yaml:"alias"`
	Path    string                 `yaml:"path"`
	Map     map[string]string      `yaml:"map"`
	Include []string               `yaml:"include"`
	Exclude []string               `yaml:"exclude"`
	Config  map[string]interface{} `yaml:"config"`
	Export  ExportType             `yaml:"export"`
	output  output.Output
}

func (m *Manifest) GetLayers(ctx AppContext) (layers []LayerConfig) {
	for _, l := range m.Layers {
		if l.Matches(ctx.Parameters) {
			layers = append(layers, l)
		}
	}
	return
}

func (l *LayerConfig) Matches(p Parameters) bool {
	// TODO: Log matching attempts
	for lpk, lpv := range l.Match {
		if pv, ok := p[lpk]; ok {
			matched, err := regexp.MatchString(lpv, pv)
			if err != nil {
				// TODO: Log error
				if lpv == pv {
					return true
				}
			}
			if !matched {
				return false
			}
		} else {
			return false
		}
	}
	return true
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
