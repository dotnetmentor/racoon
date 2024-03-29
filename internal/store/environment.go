package store

import (
	"fmt"
	"os"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/utils"
	"github.com/joho/godotenv"
)

func newEnvironment() (*Environment, error) {
	return &Environment{
		dotfilesLoaded: make([]string, 0),
	}, nil
}

type Environment struct {
	dotfilesLoaded []string
}

func (s *Environment) Read(ctx config.AppContext, layer api.Layer, key string, sensitive bool, propertySource config.ValueFromEnvironment, sourceConfig config.EnvConfig) api.Value {
	for _, dff := range sourceConfig.Dotfiles {
		df := ctx.Replace(dff)
		if utils.StringSliceContains(s.dotfilesLoaded, df) {
			continue
		}

		if err := godotenv.Overload(df); err != nil {
			if os.IsNotExist(err) {
				ctx.Log.Warnf("dotenv file %s was not found", df)
				s.dotfilesLoaded = append(s.dotfilesLoaded, df)
				continue
			} else {
				return api.NewValue(api.NewValueSource(layer, api.SourceTypeEnvironment), "", "", err, sensitive)
			}
		}

		ctx.Log.Debugf("dotenv file %s loaded", df)
		s.dotfilesLoaded = append(s.dotfilesLoaded, df)
	}

	keys := make([]string, 0)

	if len(propertySource.Key) > 0 {
		keys = append(keys, propertySource.Key)
	} else {
		keys = append(keys, key)
		keys = append(keys, utils.FormatKey(key, utils.Formatting{
			Uppercase:     true,
			WordSeparator: "_",
			PathSeparator: "_",
		}))
	}

	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			return api.NewValue(api.NewValueSource(layer, api.SourceTypeEnvironment), k, v, nil, sensitive)
		}
	}

	return api.NewValue(api.NewValueSource(layer, api.SourceTypeEnvironment), fmt.Sprintf("%v", keys), "", api.NewNotFoundError(nil, fmt.Sprintf("%v", keys), api.SourceTypeEnvironment), sensitive)
}
