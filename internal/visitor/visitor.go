package visitor

import (
	"fmt"

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
	api.SetLogger(vs.context.Log)

	vs.context.Log.Debugf("initializing visitor")
	implicit := config.PropertyList{}

	base, err := api.NewLayer("base", []config.SourceType{}, vs.context.Manifest.Config.Sources, true)
	if err != nil {
		return err
	}
	explicit := vs.context.Manifest.Properties.Filter(excludes, includes)
	vs.loadProperties(&base, implicit, explicit)
	implicit = explicit.Merge(implicit)
	vs.layers = append(vs.layers, base)

	if len(vs.context.Manifest.Config.Parameters) > 0 {
		vs.context.Log.Infof("matching layers with parameters (%s)", vs.context.Parameters.String())
	}
	ls, err := vs.context.Manifest.GetLayers(vs.context)
	if err != nil {
		return err
	}

	for _, l := range ls {
		layer, err := api.NewLayer(l.Name, l.ImplicitSources, l.Config, false)
		if err != nil {
			return err
		}
		explicit := l.Properties.Filter(excludes, includes)
		vs.loadProperties(&layer, implicit, explicit)
		implicit = explicit.Merge(implicit)
		vs.layers = append(vs.layers, layer)
	}

	vs.context.Log.Debug("visitor initialized")
	return nil
}

func (vs *Visitor) Store() *store.ValueStore {
	return vs.store
}

func (vs *Visitor) Layer(action func(l api.Layer, err error) (bool, error)) error {
	for _, l := range vs.layers {
		vs.context.Log.Debugf("visiting layer %s", l.Name)

		if ok, err := action(l, nil); !ok || err != nil {
			return err
		}
	}

	return nil
}

func (vs *Visitor) Property(action func(p api.Property, err error) (bool, error)) error {
	for _, p := range vs.properties {
		vs.context.Log.Debugf("visiting property %s", p.Name)

		err := vs.layers.ResolveValue(&p)
		if ok, err := action(p, err); !ok || err != nil {
			return err
		}
	}

	return nil
}

func (vs *Visitor) loadProperties(layer *api.Layer, implicit, explicit config.PropertyList) {
	vs.context.Log.Infof("processing layer %s", layer.Name)

	if len(layer.ImplicitSources) > 0 {
		for _, p := range implicit.Remove(explicit) {
			prop, _ := vs.newProperty(p.Name, p.Description, layer.Name, p.Sensitive, p.Rules, p.Format)

			if !prop.Rules().Override.AllowImplicit {
				vs.context.Log.Debugf("skipping property %s, implicit overrides are not allowed by property rules", prop.Name)
				continue
			}

			for _, s := range layer.ImplicitSources {
				vs.context.Log.Debugf("processing implicit property %s, reading from source %s", prop.Name, s)

				var valueSource *config.ValueSourceConfig = nil
				switch s {
				case config.SourceTypeAwsParameterStore:
					valueSource = &config.ValueSourceConfig{
						AwsParameterStore: &config.ValueFromAwsParameterStore{},
					}
				case config.SourceTypeEnvironment:
					valueSource = &config.ValueSourceConfig{
						Environment: &config.ValueFromEvnironment{},
					}
				default:
					vs.context.Log.Warnf("unsupported implicit source %s", s)
				}

				if valueSource != nil {
					val := vs.store.Read(*layer, prop.Name, prop.Sensitive(), valueSource, layer.Config)
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
			vs.context.Log.Warnf("skipping property %s, explicit overrides are not allowed by property rules", prop.Name)
			continue
		}

		vs.context.Log.Debugf("processing explicit property %s", prop.Name)

		if ok && p.Default != nil {
			dv := *p.Default
			vs.context.Log.Debugf("%s, setting default value to: %s", prop.Name, dv)
			prop.SetValue(api.NewValue(api.NewValueSource(*layer, api.SourceTypeDefault), "", dv, nil, p.Sensitive))
		}

		if p.Source != nil {
			val := vs.store.Read(*layer, prop.Name, prop.Sensitive(), p.Source, layer.Config)
			if val != nil {
				prop.SetValue(val)
			}
		}

		val := prop.Value()
		if len(prop.Formatting()) > 0 && val != nil {
			vs.context.Log.Debugf("formatting value for %s, format: %s", prop.Name, val.String())

			str := val.Raw()
			errs := make([]*api.FormattingError, 0)
			forceSensitive := prop.Sensitive()

			for _, fc := range prop.Formatting() {
				f := api.NewFormatter(fc, vs.context.Log)
				k := f.FormattingKey()

				fval := vs.store.Read(*layer, k, prop.Sensitive(), fc.Source, layer.Config)
				if fval != nil {
					if fval.Error() != nil {
						msg := fmt.Sprintf("failed to read %s value from %s, used to format %s, err: %v", k, fval.Source(), prop.String(), fval.Error())
						vs.context.Log.Debugln(msg)
						errs = append(errs, api.NewFormattingError(msg))
						continue
					}

					if fval.Sensitive() {
						forceSensitive = true
					}

					vs.context.Log.Debugf("applying formatter for %s using value %s (source=%s formatter=%s)", prop.Name, fval.String(), fval.Source().Type(), f.String())
					res, err := f.Apply(str, fval)
					// TODO: Verify sensitive values can't be part of the error string
					if err != nil {
						msg := fmt.Sprintf("failed to apply formatting for %s using %T, err: %v", k, f, err)
						vs.context.Log.Debugln(msg)
						errs = append(errs, api.NewFormattingError(msg))
						continue
					}

					str = res
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

func (vs *Visitor) newProperty(name, description string, source string, sensitive bool, rules config.RuleConfig, formatting []config.FormattingConfig) (property api.Property, isNew bool) {
	property, isNew = api.NewProperty(vs.properties, name, description, source, sensitive, rules, formatting)
	if isNew {
		vs.properties = append(vs.properties, property)
	}
	return
}
