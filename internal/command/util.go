package command

import (
	"fmt"

	"github.com/dotnetmentor/racoon/internal/backend"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func manifestPaths(c *cli.Context) []string {
	paths := config.DefaultManifestYamlFiles
	manifest := c.String("manifest")
	if manifest != "" {
		paths = []string{manifest}
	}
	return paths
}

func newContext(c *cli.Context, metadata config.AppMetadata, validateParams bool) (config.AppContext, error) {
	paths := manifestPaths(c)
	ctx, err := config.NewContext(metadata, paths...)
	if err != nil {
		ctx.Log.Error(err)
		ctx.Log.Exit(1)
	}

	l := c.String("loglevel")
	level, err := logrus.ParseLevel(l)
	if err != nil {
		ctx.Log.Error(err)
		ctx.Log.Exit(1)
	}
	ctx.Log.SetLevel(level)

	p, err := config.ParseParams(c.StringSlice("parameter"))
	if err != nil {
		ctx.Log.Error(err)
		ctx.Log.Exit(1)
	}

	if validateParams {
		if err := p.ValidateParams(ctx.Manifest.Config.Parameters); err != nil {
			return ctx, err
		}
	}

	ctx.Parameters = p.Ordered(ctx.Manifest.Config.Parameters)
	ctx.Context = c.Context

	return ctx, nil
}

func newBackend(ctx config.AppContext) (backend.Backend, error) {
	if ctx.Manifest.Backend.Enabled {
		if ctx.Manifest.Name == "" {
			return nil, fmt.Errorf("manifest name must be set in order to use backend")
		}
		backend, err := backend.New(ctx.Context, ctx.Manifest.Backend)
		if err != nil {
			return nil, err
		}
		return backend, nil
	}
	return nil, nil
}
