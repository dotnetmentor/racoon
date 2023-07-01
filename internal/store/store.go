package store

import (
	"fmt"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/config"
)

func NewValueStore(ctx config.AppContext) *ValueStore {
	return &ValueStore{
		context: ctx,
	}
}

type ValueStore struct {
	context config.AppContext

	awsParameterStore *AwsParameterStore
	environment       *Environment
}

func (vs *ValueStore) Read(layer api.Layer, key string, sensitive bool, source *config.PropertyValueFrom, sourceConfig config.SourceConfig) api.Value {
	m := vs.context.Manifest

	switch source.SourceType() {
	case config.SourceTypeParameter:
		if len(*source.Parameter) > 0 {
			key := *source.Parameter
			if v, ok := vs.context.Parameters[key]; ok {
				return api.NewValue(api.NewValueSource(layer, api.SourceTypeParameter), key, v, nil, sensitive)
			} else {
				return api.NewValue(api.NewValueSource(layer, api.SourceTypeParameter), key, v, api.NewNotFoundError(nil, key, api.SourceTypeParameter), sensitive)
			}
		}

	case config.SourceTypeLiteral:
		return api.NewValue(api.NewValueSource(layer, api.SourceTypeLiteral), key, *source.Literal, nil, sensitive)

	case config.SourceTypeEnvironment:
		if vs.environment == nil {
			store, err := newEnvironment()
			if err != nil {
				return api.NewValue(api.NewValueSource(layer, api.SourceTypeEnvironment), "", "", err, sensitive)
			}
			vs.environment = store
		}

		return vs.environment.Read(vs.context, layer, key, sensitive, *source.Environment, sourceConfig.Env)

	case config.SourceTypeAwsParameterStore:
		mc := m.Config.Sources.AwsParameterStore.Merge(sourceConfig.AwsParameterStore)
		if vs.awsParameterStore == nil {
			store, err := newAwsParameterStore(vs.context.Context)
			if err != nil {
				return api.NewValue(api.NewValueSource(layer, api.SourceTypeAwsParameterStore), "", "", err, sensitive || mc.ForceSensitive)
			}
			vs.awsParameterStore = store
		}

		return vs.awsParameterStore.Read(vs.context, layer, key, sensitive, *source.AwsParameterStore, mc)
	}

	return nil
}

func (vs *ValueStore) Write(key, value, description string, sourceType api.SourceType, sourceConfig config.SourceConfig) error {
	if !sourceType.Writable() {
		return fmt.Errorf("unsupported source type %s, source is not writable", sourceType)
	}

	m := vs.context.Manifest

	switch sourceType {
	case api.SourceTypeAwsParameterStore:
		mc := m.Config.Sources.AwsParameterStore.Merge(sourceConfig.AwsParameterStore)
		return vs.awsParameterStore.Write(vs.context, key, value, description, mc)
	}

	return nil
}
