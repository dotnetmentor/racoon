package api

import (
	"fmt"

	"github.com/dotnetmentor/racoon/internal/config"
)

func NewLayer(name string, implicitSources []config.SourceType, sourceConfig config.SourceConfig, baseLayer bool) (Layer, error) {
	is := make(map[config.SourceType]struct{})

	for _, s := range implicitSources {
		if _, ok := is[s]; ok {
			return Layer{}, NewConfigurationError(fmt.Sprintf("implicit sources must be unique, %s found multiple times in layer %s", s, name))
		}
		is[s] = struct{}{}
	}

	l := Layer{
		Name:            name,
		Properties:      make([]Property, 0),
		ImplicitSources: implicitSources,
		Config:          sourceConfig,
		baseLayer:       baseLayer,
	}
	return l, nil
}

type LayerList []Layer

func (ls LayerList) ResolveValue(p *Property) (err error) {
	for _, l := range ls {
		lp := l.Property(p.Name)
		if lp != nil {
			for _, v := range lp.Values() {
				p.SetValue(v)
			}
		}
	}
	return
}

type Layer struct {
	Name            string
	Properties      PropertyList
	ImplicitSources []config.SourceType
	Config          config.SourceConfig
	baseLayer       bool
}

func (l Layer) IsBaseLayer() bool {
	return l.baseLayer
}

func (l Layer) Property(property string) *Property {
	for _, p := range l.Properties {
		if p.Name == property {
			return &p
		}
	}
	return nil
}

func (l Layer) Value(property string) Value {
	p := l.Property(property)
	if p != nil {
		return p.Value()
	}
	return nil
}
