package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/dotnetmentor/racoon/internal/aws"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/utils"
	"github.com/joho/godotenv"
)

type ValueSource struct {
	context    config.AppContext
	properties PropertyList

	dotfilesLoaded    []string
	awsParameterStore *ssm.Client
}

type LayerList []Layer

type Layer struct {
	Name            string
	ImplicitSources []config.SourceType
	Properties      PropertyList
	baseLayer       bool
}

func NewLayer(name string, implicitSources []config.SourceType, baseLayer bool) Layer {
	l := Layer{
		Name:            name,
		Properties:      make([]Property, 0),
		ImplicitSources: implicitSources,
		baseLayer:       baseLayer,
	}
	return l
}

type PropertyList []Property

type Property struct {
	Name        string
	Description string

	source    string
	value     Value
	sensitive bool
	rules     config.RuleConfig
}

type Value interface {
	Source() string
	Raw() string
	Error() error
	String() string
}

type SensitiveValue struct {
	source string
	raw    string
	err    error
}

func (v *SensitiveValue) Source() string {
	return v.source
}

func (v *SensitiveValue) Raw() string {
	return v.raw
}

func (v *SensitiveValue) Error() error {
	return v.err
}

func (v *SensitiveValue) String() string {
	return "<sensitive>"
}

type ClearTextValue struct {
	source string
	raw    string
	err    error
}

func (v *ClearTextValue) Source() string {
	return v.source
}

func (v *ClearTextValue) Raw() string {
	return v.raw
}

func (v *ClearTextValue) Error() error {
	return v.err
}

func (v *ClearTextValue) String() string {
	if len(v.raw) == 0 {
		return "<empty>"
	}
	return v.raw
}

func (ls LayerList) ResolveValue(p *Property) (err error) {
	for _, l := range ls {
		val := l.Value(p.Name)
		if val != nil {
			err = val.Error()
			if err != nil {
				return
			}
			if p.value == nil || p.value.Raw() != val.Raw() {
				p.value = val
			}
		}
	}
	return
}

func (l Layer) Value(property string) Value {
	for _, p := range l.Properties {
		if p.Name == property {
			return p.Value()
		}
	}
	return nil
}

func (p *Property) SetValue(layer Layer, source config.SourceType, val string, err error, forceSensitive bool) {
	p.value = NewValue(layer.SourceName(source), val, err, p.sensitive || forceSensitive)
}

func NewValue(source, val string, err error, sensitive bool) Value {
	if sensitive {
		return &SensitiveValue{
			source: source,
			raw:    val,
			err:    err,
		}
	} else {
		return &ClearTextValue{
			source: source,
			raw:    val,
			err:    err,
		}
	}
}

func (p *Property) Value() Value {
	return p.value
}

func (p *Property) String() string {
	return fmt.Sprintf("%s/%s", p.source, p.Name)
}

func (vs *ValueSource) NewProperty(name, description, source string, sensitive bool, rules config.RuleConfig) (property Property, ok bool) {
	property = Property{
		Name:        name,
		Description: description,
		source:      source,
		sensitive:   sensitive,
		rules:       rules,
	}

	found := false
	for _, p := range vs.properties {
		if p.Name == property.Name {
			found = true
			if !property.sensitive {
				property.sensitive = p.sensitive
			}
			property.rules = p.rules
			break
		}
	}

	if !found {
		vs.properties = append(vs.properties, property)
		ok = true
	}
	return
}

func (vs *ValueSource) LoadProperties(layer *Layer, implicit, explicit config.PropertyList, sourceConfig config.SourceConfig) {
	vs.context.Log.Infof("processing %s layer", layer.Name)

	if len(layer.ImplicitSources) > 0 {
		for _, p := range implicit.Remove(explicit) {
			prop, _ := vs.NewProperty(p.Name, p.Description, layer.Name, p.Sensitive, p.Rules)

			if prop.rules.Override.DenyImplicit {
				vs.context.Log.Debugf("skipping property %s as implicit overrides are denied by property rules", prop.Name)
				continue
			}

			for _, s := range layer.ImplicitSources {
				vs.context.Log.Debugf("processing implicit property %s, reading from source %s", prop.Name, s)

				var valueSource *config.PropertyValueFrom = nil
				switch s {
				case config.SourceTypeAwsParameterStore:
					valueSource = &config.PropertyValueFrom{
						AwsParameterStore: &config.ValueFromAwsParameterStoreConfig{},
					}
				case config.SourceTypeEnvironment:
					valueSource = &config.PropertyValueFrom{
						Environment: &config.ValueFromEvnironment{},
					}
				default:
					vs.context.Log.Warnf("unsupported implicit source %s", s)
				}

				if valueSource != nil {
					val := vs.readFromSource(*layer, prop.Name, prop.sensitive, valueSource, sourceConfig)
					if val != nil {
						prop.value = val
					}
				}
			}

			layer.Properties = append(layer.Properties, prop)
		}
	}

	for _, p := range explicit {
		prop, ok := vs.NewProperty(p.Name, p.Description, layer.Name, p.Sensitive, p.Rules)
		if !layer.IsBaseLayer() && prop.rules.Override.DenyExplicit {
			vs.context.Log.Debugf("skipping property %s as explicit overrides are denied by property rules", prop.Name)
			continue
		}

		vs.context.Log.Debugf("processing explicit property %s", prop.Name)

		if ok && p.Default != nil {
			dv := *p.Default
			vs.context.Log.Debugf("%s, setting default value to: %s", prop.Name, dv)
			prop.value = &ClearTextValue{
				source: "default",
				raw:    dv,
			}
		}

		if p.Source != nil {
			val := vs.readFromSource(*layer, prop.Name, prop.sensitive, p.Source, sourceConfig)
			if val != nil {
				prop.value = val
			}
		}

		val := prop.Value()
		if p.Format != nil && val != nil {
			vs.context.Log.Debugf("formatting value for %s, format: %s", prop.Name, val.String())

			rval := val.Raw()
			forceSensitive := prop.sensitive
			for k, v := range p.Format.Replace {
				nval := vs.readFromSource(*layer, k, prop.sensitive, v, sourceConfig)
				rkey := fmt.Sprintf("{%s}", k)
				if nval != nil {
					if nval.Error() != nil {
						vs.context.Log.Errorf("failed to read %s value from %s, used to format %s, err: %v", k, nval.Source(), prop.String(), nval.Error())
						continue
					}

					if nval.Raw() == "" {
						vs.context.Log.Errorf("failed to read %s value from %s, used to format %s, err: empty value not allowed", k, prop.String(), nval.Source())
						continue
					}

					switch nval.(type) {
					case *SensitiveValue:
						forceSensitive = true
					}

					rval = strings.ReplaceAll(rval, rkey, nval.Raw())
					vs.context.Log.Debugf("replaced %s with value: %s", rkey, nval.String())
				} else {
					vs.context.Log.Errorf("failed to read %s value, used to format %s, err: no value received", k, prop.String())
					continue
				}
			}

			if val.Raw() != rval {
				prop.SetValue(*layer, "valueformatter", rval, nil, forceSensitive)
			}
		}

		layer.Properties = append(layer.Properties, prop)
	}
}

func (vs *ValueSource) ReadOne(key string) (value Value, err error) {
	_, values, err := vs.ReadAll([]string{}, []string{key})
	if err != nil {
		return nil, err
	}
	return values[key], nil
}

func (vs *ValueSource) ReadAll(excludes, includes []string) (keys []string, values map[string]Value, err error) {
	keys = []string{}
	values = map[string]Value{}
	layers := LayerList{}
	implicit := config.PropertyList{}

	base := NewLayer("base", []config.SourceType{}, true)
	explicit := vs.context.Manifest.Properties.Filter(excludes, includes)
	vs.LoadProperties(&base, implicit, explicit, vs.context.Manifest.Config.Sources)
	implicit = explicit.Merge(implicit)
	layers = append(layers, base)

	for _, l := range vs.context.Manifest.GetLayers(vs.context) {
		layer := NewLayer(l.Name, l.ImplicitSources, false)
		explicit := l.Properties.Filter(excludes, includes)
		vs.LoadProperties(&layer, implicit, explicit, l.Config)
		implicit = explicit.Merge(implicit)
		layers = append(layers, layer)
	}

	for _, p := range vs.properties {
		key := p.Name

		if !utils.StringSliceContains(keys, key) {
			keys = append(keys, key)
		}

		err := layers.ResolveValue(&p)
		if err != nil {
			return nil, nil, err
		}

		val := p.Value()
		if val == nil {
			return nil, nil, fmt.Errorf("no value resolved for property %s", p.Name)
		}

		if val.Error() != nil {
			return nil, nil, val.Error()
		}

		if len(val.Raw()) == 0 && !p.rules.Validation.AllowEmpty {
			return nil, nil, fmt.Errorf("empty value not allowed for property %s (source=%s)", p.Name, val.Source())
		}

		values[key] = val

		vs.context.Log.Debugf("property %s, defined in %s, value from %s, value set to: %s", p.Name, p.source, val.Source(), val.String())
	}

	return keys, values, nil
}

func (l Layer) SourceName(s config.SourceType) string {
	return fmt.Sprintf("%s/%s", l.Name, s)
}

func (l Layer) IsBaseLayer() bool {
	return l.baseLayer
}

func (vs *ValueSource) readFromSource(layer Layer, key string, sensitive bool, source *config.PropertyValueFrom, sourceConfig config.SourceConfig) Value {
	m := vs.context.Manifest

	if source != nil {
		if source.Parameter != nil && len(*source.Parameter) > 0 {
			key := *source.Parameter
			if v, ok := vs.context.Parameters[key]; ok {
				return NewValue(layer.SourceName(config.SourceTypeParameter), v, nil, sensitive)
			}
			// TODO: If parameter "key" is not found, should probably be handled as an error
		}

		if source.Literal != nil {
			return NewValue(layer.SourceName(config.SourceTypeLiteral), *source.Literal, nil, sensitive)
		}

		if source.Environment != nil {
			for _, dff := range sourceConfig.Env.Dotfiles {
				df := replaceParams(dff, vs.context.Parameters)
				if utils.StringSliceContains(vs.dotfilesLoaded, df) {
					continue
				}
				vs.dotfilesLoaded = append(vs.dotfilesLoaded, df)

				if err := godotenv.Overload(df); err != nil {
					if os.IsNotExist(err) {
						vs.context.Log.Infof("dotenv file %s was not found", df)
					} else {
						return NewValue(layer.SourceName(config.SourceTypeEnvironment), "", err, sensitive)
					}
				}
			}

			keys := make([]string, 0)

			if len(source.Environment.Key) > 0 {
				keys = append(keys, source.Environment.Key)
			} else {
				keys = append(keys, key)
				keys = append(keys, utils.CamelCaseSplitToUpperJoinByUnderscore(key))
			}

			for _, k := range keys {
				if v, ok := os.LookupEnv(k); ok {
					return NewValue(layer.SourceName(config.SourceTypeEnvironment), v, nil, sensitive)
				}
			}
		}

		if source.AwsParameterStore != nil {
			mc := m.Config.Sources.AwsParameterStore.Merge(sourceConfig.AwsParameterStore)
			if vs.awsParameterStore == nil {
				client, err := aws.NewParameterStoreClient(vs.context.Context)
				if err != nil {
					return NewValue(layer.SourceName(config.SourceTypeAwsParameterStore), "", err, sensitive || mc.ForceSensitive)
				}
				vs.awsParameterStore = client
			}

			pskf := mc.KeyFormat
			if len(source.AwsParameterStore.Key) > 0 {
				pskf = source.AwsParameterStore.Key
			}

			psk := aws.ParameterStoreKey(replaceParams(pskf, vs.context.Parameters), key)
			vs.context.Log.Debugf("reading %s from %s", psk, config.SourceTypeAwsParameterStore)
			out, err := vs.awsParameterStore.GetParameter(vs.context.Context, &ssm.GetParameterInput{
				Name:           &psk,
				WithDecryption: true,
			})
			if err != nil {
				var notFound *ssmtypes.ParameterNotFound
				if !errors.As(err, &notFound) {
					return NewValue(layer.SourceName(config.SourceTypeAwsParameterStore), "", err, sensitive || mc.ForceSensitive)
				}
				vs.context.Log.Debugf("%s not found in %s", psk, config.SourceTypeAwsParameterStore)
			} else {
				return NewValue(layer.SourceName(config.SourceTypeAwsParameterStore), *out.Parameter.Value, err, sensitive || mc.ForceSensitive)
			}
		}
	}

	return nil
}
