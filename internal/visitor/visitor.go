package visitor

import (
	"fmt"
	"strings"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/store"
)

func New(ctx config.AppContext) *Visitor {
	v := &Visitor{
		context:    ctx,
		store:      store.NewValueStore(ctx),
		properties: make(api.PropertyList, 0),
		layers:     make(api.LayerList, 0),
	}
	return v
}

type Visitor struct {
	context    config.AppContext
	store      *store.ValueStore
	properties api.PropertyList
	layers     api.LayerList
}

func (vs *Visitor) Init(excludes, includes []string) error {
	vs.context.Log.Debugf("initializing visitor")
	implicit := config.PropertyList{}

	base := api.NewLayer("base", []config.SourceType{}, true)
	explicit := vs.context.Manifest.Properties.Filter(excludes, includes)
	vs.loadProperties(&base, implicit, explicit, vs.context.Manifest.Config.Sources)
	implicit = explicit.Merge(implicit)
	vs.layers = append(vs.layers, base)

	for _, l := range vs.context.Manifest.GetLayers(vs.context) {
		layer := api.NewLayer(l.Name, l.ImplicitSources, false)
		explicit := l.Properties.Filter(excludes, includes)
		vs.loadProperties(&layer, implicit, explicit, l.Config)
		implicit = explicit.Merge(implicit)
		vs.layers = append(vs.layers, layer)
	}

	vs.context.Log.Debug("visitor initialized")
	return nil
}

func (vs *Visitor) Store() *store.ValueStore {
	return vs.store
}

func (vs *Visitor) Property(action func(p api.Property, err error) error) error {
	for _, p := range vs.properties {
		vs.context.Log.Debugf("visiting property %s", p.Name)

		err := vs.layers.ResolveValue(&p)
		if err := action(p, err); err != nil {
			return err
		}
	}

	return nil
}

func (vs *Visitor) loadProperties(layer *api.Layer, implicit, explicit config.PropertyList, sourceConfig config.SourceConfig) {
	vs.context.Log.Infof("processing layer %s", layer.Name)

	if len(layer.ImplicitSources) > 0 {
		for _, p := range implicit.Remove(explicit) {
			prop, _ := vs.newProperty(p.Name, p.Description, layer.Name, p.Sensitive, p.Rules, p.Format)

			if !prop.Rules().Override.AllowImplicit {
				vs.context.Log.Debugf("skipping property %s as implicit overrides are denied by property rules", prop.Name)
				continue
			}

			for _, s := range layer.ImplicitSources {
				vs.context.Log.Debugf("processing implicit property %s, reading from source %s", prop.Name, s)

				var valueSource *config.PropertyValueFrom = nil
				switch s {
				case config.SourceTypeAwsParameterStore:
					valueSource = &config.PropertyValueFrom{
						AwsParameterStore: &config.ValueFromAwsParameterStore{},
					}
				case config.SourceTypeEnvironment:
					valueSource = &config.PropertyValueFrom{
						Environment: &config.ValueFromEvnironment{},
					}
				default:
					vs.context.Log.Warnf("unsupported implicit source %s", s)
				}

				if valueSource != nil {
					val := vs.store.Read(*layer, prop.Name, prop.Sensitive(), valueSource, sourceConfig)
					if val != nil {
						prop.SetValue(val)
					}
				}
			}

			layer.Properties = append(layer.Properties, prop)
		}
	}

	for _, p := range explicit {
		prop, ok := vs.newProperty(p.Name, p.Description, layer.Name, p.Sensitive, p.Rules, p.Format)
		if !layer.IsBaseLayer() && !prop.Rules().Override.AllowExplicit {
			vs.context.Log.Debugf("skipping property %s as explicit overrides are denied by property rules", prop.Name)
			continue
		}

		vs.context.Log.Debugf("processing explicit property %s", prop.Name)

		if ok && p.Default != nil {
			dv := *p.Default
			vs.context.Log.Debugf("%s, setting default value to: %s", prop.Name, dv)
			prop.SetValue(api.NewValue(api.NewValueSource(*layer, api.SourceTypeDefault), "", dv, nil, false))
		}

		if p.Source != nil {
			val := vs.store.Read(*layer, prop.Name, prop.Sensitive(), p.Source, sourceConfig)
			if val != nil {
				prop.SetValue(val)
			}
		}

		val := prop.Value()
		if prop.Formatting() != nil && val != nil {
			vs.context.Log.Debugf("formatting value for %s, format: %s", prop.Name, val.String())

			errs := make([]*api.FormattingError, 0)
			str := val.Raw()
			forceSensitive := prop.Sensitive()
			for k, v := range prop.Formatting().Replace {
				fval := vs.store.Read(*layer, k, prop.Sensitive(), v, sourceConfig)
				rkey := fmt.Sprintf("{%s}", k)
				if fval != nil {
					if fval.Error() != nil {
						vs.context.Log.Debugf("failed to read %s value from %s, used to format %s, err: %v", k, fval.Source(), prop.String(), fval.Error())
						errs = append(errs, api.NewFormattingError(fmt.Sprintf("failed to read %s value from %s, used to format %s, err: %v", k, fval.Source(), prop.String(), fval.Error())))
						continue
					}

					if fval.Raw() == "" {
						vs.context.Log.Debugf("failed to read %s value from %s, used to format %s, err: empty value not allowed", k, prop.String(), fval.Source())
						errs = append(errs, api.NewFormattingError(fmt.Sprintf("failed to read %s value from %s, used to format %s, err: empty value not allowed", k, prop.String(), fval.Source())))
						continue
					}

					if fval.Sensitive() {
						forceSensitive = true
					}

					str = strings.ReplaceAll(str, rkey, fval.Raw())
					vs.context.Log.Debugf("replaced %s with value: %s", rkey, fval.String())
				} else {
					vs.context.Log.Debugf("failed to read %s value, used to format %s, err: no value received", k, prop.String())
					errs = append(errs, api.NewFormattingError(fmt.Sprintf("failed to read %s value, used to format %s, err: no value received", k, prop.String())))
					continue
				}
			}

			err := api.WrapFormattingErrors(errs)
			prop.SetRawValue(*layer, api.SourceTypeFormatter, "", str, err, forceSensitive)
		}

		layer.Properties = append(layer.Properties, prop)
	}
}

func (vs *Visitor) newProperty(name, description string, source string, sensitive bool, rules config.RuleConfig, formatting *config.FormattingConfig) (property api.Property, isNew bool) {
	property, isNew = api.NewProperty(vs.properties, name, description, source, sensitive, rules, formatting)
	if isNew {
		vs.properties = append(vs.properties, property)
	}
	return
}
