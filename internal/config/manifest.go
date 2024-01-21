package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dotnetmentor/racoon/internal/output"
	"github.com/dotnetmentor/racoon/internal/utils"

	yaml2 "gopkg.in/yaml.v2"
)

const (
	SourceTypeNotSet            SourceType = "unknown"
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

var (
	DefaultPropertyRules RuleConfig = RuleConfig{
		Validation: ValidationRuleConfig{
			AllowEmpty: false,
		},
		Override: OverrideRuleConfig{
			AllowImplicit: true,
			AllowExplicit: true,
		},
	}
)

type SourceType string

type OutputType string

type ExportType string

func NewManifest(paths []string) (Manifest, error) {
	// base path
	basepath, _ := os.Getwd()

	// read manifest
	m, err := readManifest(basepath, paths)
	if err != nil {
		return m, err
	}

	// TODO: Validate manifest config
	layers := make(map[string]interface{})
	for _, l := range m.Layers {
		if _, ok := layers[l.Name]; ok {
			return m, fmt.Errorf("duplicate layer, %s defined multiple times", l.Name)
		}
		layers[l.Name] = nil
	}

	return m, nil
}

func readManifest(basepath string, paths []string) (Manifest, error) {
	// read manifest file
	var file []byte
	var path string

	for _, filename := range paths {
		var fullpath string
		if filepath.IsAbs(filename) {
			fullpath = filename
		} else {
			fullpath = filepath.Join(basepath, filename)
		}

		if _, err := os.Stat(fullpath); os.IsNotExist(err) {
			continue
		}

		bs, err := os.ReadFile(fullpath)
		if err != nil {
			return Manifest{}, fmt.Errorf("failed to read manifest file (path=%s). %v", fullpath, err)
		}
		file = bs
		path = fullpath
		break
	}

	if file == nil {
		return Manifest{}, fmt.Errorf("failed to find manifest file paths=%v", paths)
	}

	// parse base config
	ec := ExtendsConfig{}
	if err := yaml2.Unmarshal(file, &ec); err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest base yaml (%s), %v", path, err)
	}

	// parse manifest
	m := Manifest{}

	if len(ec.Extends) > 0 {
		bm, err := readManifest(filepath.Dir(path), []string{ec.Extends})
		if err != nil {
			return Manifest{}, err
		}
		m = bm
	}

	m.filepath = path

	if err := yaml2.UnmarshalStrict(file, &m); err != nil {
		return Manifest{}, fmt.Errorf("failed to parse manifest yaml (%s), %v", path, err)
	}

	return m, nil
}

type Manifest struct {
	filepath       string
	ExtendsConfig  `yaml:",inline"`
	MetadataConfig `yaml:",inline"`
	Config         Config         `yaml:"config"`
	Layers         LayerList      `yaml:"layers"`
	Properties     PropertyList   `yaml:"properties"`
	Outputs        []OutputConfig `yaml:"outputs"`
}

type ExtendsConfig struct {
	Extends string `yaml:"extends"`
}

type MetadataConfig struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels"`
}

func (m Manifest) Filepath() string {
	return m.filepath
}

type Config struct {
	Parameters ParameterConfigList `yaml:"parameters"`
	Sources    SourceConfig        `yaml:"sources"`
}

type LayerList []LayerConfig

func (s *LayerList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawLayerList LayerList

	raw := rawLayerList{}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	ll := LayerList(raw)
	ll = append(*s, ll...)
	*s = ll

	return nil
}

type LayerConfig struct {
	Name            string       `yaml:"name"`
	Match           []string     `yaml:"match"`
	Config          SourceConfig `yaml:"config"`
	ImplicitSources []SourceType `yaml:"implicitSources"`
	Properties      PropertyList `yaml:"properties"`
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

type ParameterConfigList []ParameterConfig

func (p ParameterConfigList) HasKey(k string) bool {
	for _, r := range p {
		if r.Key == k {
			return true
		}
	}
	return false
}

type ParameterConfig struct {
	Key      string `yaml:"key"`
	Required bool   `yaml:"required"`
	Regexp   string `yaml:"regexp"`
}

type SourceConfig struct {
	AwsParameterStore AwsParameterStoreConfig `yaml:"awsParameterStore"`
	Env               EnvConfig               `yaml:"env"`
}

type AwsParameterStoreConfig struct {
	DefaultKey           string `yaml:"defaultKey"`
	KmsKey               string `yaml:"kmsKey"`
	ForceSensitive       bool   `yaml:"forceSensitive"`
	TreatNotFoundAsError bool   `yaml:"treatNotFoundAsError"`
}

func (c AwsParameterStoreConfig) Merge(config AwsParameterStoreConfig) AwsParameterStoreConfig {
	nc := AwsParameterStoreConfig{
		DefaultKey:           c.DefaultKey,
		ForceSensitive:       c.ForceSensitive,
		KmsKey:               c.KmsKey,
		TreatNotFoundAsError: c.TreatNotFoundAsError,
	}

	if len(config.DefaultKey) > 0 && nc.DefaultKey != config.DefaultKey {
		nc.DefaultKey = config.DefaultKey
	}

	if config.ForceSensitive {
		nc.ForceSensitive = true
	}

	if len(config.KmsKey) > 0 && nc.KmsKey != config.KmsKey {
		nc.KmsKey = config.KmsKey
	}

	if config.TreatNotFoundAsError {
		nc.TreatNotFoundAsError = true
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
	Source      *ValueSourceConfig `yaml:"source,omitempty"`
	Format      []FormattingConfig `yaml:"format,omitempty"`
	Rules       RuleConfig         `yaml:"rules"`
}

func (s *PropertyConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawConfig PropertyConfig

	// Change defaults in DefaultPropertyRules
	raw := rawConfig{
		Rules: DefaultPropertyRules,
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	*s = PropertyConfig(raw)
	return nil
}

type FormattingConfig struct {
	Replace       *string            `yaml:"replace,omitempty"`
	RegexpReplace *string            `yaml:"regexpReplace,omitempty"`
	Source        *ValueSourceConfig `yaml:"source,omitempty"`
}

type RuleConfig struct {
	Validation ValidationRuleConfig `yaml:"validation"`
	Override   OverrideRuleConfig   `yaml:"override"`
}

type ValidationRuleConfig struct {
	Optional   bool `yaml:"optional"`
	AllowEmpty bool `yaml:"allowEmpty"`
}

type OverrideRuleConfig struct {
	AllowImplicit bool `yaml:"allowImplicit"`
	AllowExplicit bool `yaml:"allowExplicit"`
}

type ValueSourceConfig struct {
	Parameter         *string                     `yaml:"parameter,omitempty"`
	Literal           *string                     `yaml:"literal,omitempty"`
	Environment       *ValueFromEvnironment       `yaml:"env,omitempty"`
	AwsParameterStore *ValueFromAwsParameterStore `yaml:"awsParameterStore,omitempty"`
}

func (s *ValueSourceConfig) SourceType() SourceType {
	if s != nil {
		if s.Parameter != nil {
			return SourceTypeParameter
		}

		if s.Literal != nil {
			return SourceTypeLiteral
		}

		if s.Environment != nil {
			return SourceTypeEnvironment
		}

		if s.AwsParameterStore != nil {
			return SourceTypeAwsParameterStore
		}
	}
	return SourceTypeNotSet
}

type ValueFromEvnironment struct {
	Key string `yaml:"key"`
}

type ValueFromAwsParameterStore struct {
	Key                  string `yaml:"key"`
	TreatNotFoundAsError *bool  `yaml:"treatNotFoundAsError"`
}

type OutputConfig struct {
	Type    OutputType             `yaml:"type"`
	Alias   string                 `yaml:"alias"`
	Paths   []string               `yaml:"paths"`
	Map     map[string]string      `yaml:"map"`
	Include []string               `yaml:"include"`
	Exclude []string               `yaml:"exclude"`
	Config  map[string]interface{} `yaml:"config"`
	Export  ExportType             `yaml:"export"`
	output  output.Output
}

func (m *Manifest) GetLayers(ctx AppContext) (layers []LayerConfig, err error) {
	for _, l := range m.Layers {
		match, err := l.Matches(ctx.Parameters, ctx)
		if err != nil {
			return layers, err
		}
		if match {
			layers = append(layers, l)
		}
	}
	return
}

func (l *LayerConfig) Matches(op OrderedParameterList, ctx AppContext) (match bool, err error) {
	match = true

	for _, expr := range l.Match {
		k, m, e := ParseExpression(expr)
		if e != nil {
			match = false
			err = fmt.Errorf("matching layer %s against parameters failed, %v", l.Name, e)
			break
		}
		if pv, ok := op.Value(k); ok {
			if !m.Match(pv) {
				match = false
				break
			}
		} else {
			match = false
			break
		}
	}

	if match {
		ctx.Log.Debugf("matched layer %s against parameters (conditions=%v parameters=%v)", l.Name, l.Match, op)
	} else {
		ctx.Log.Debugf("layer %s did not match parameters (conditions=%v parameters=%v)", l.Name, l.Match, op)
	}

	return
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
