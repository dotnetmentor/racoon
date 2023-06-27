package api

import (
	"github.com/dotnetmentor/racoon/internal/config"
)

func NewLayer(name string, implicitSources []config.SourceType, baseLayer bool) Layer {
	l := Layer{
		Name:            name,
		Properties:      make([]Property, 0),
		ImplicitSources: implicitSources,
		baseLayer:       baseLayer,
	}
	return l
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
	ImplicitSources []config.SourceType
	Properties      PropertyList
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
