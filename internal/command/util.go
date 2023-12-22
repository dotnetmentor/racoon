package command

import (
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func newContext(c *cli.Context, validateParams bool) (config.AppContext, error) {
	paths := config.DefaultManifestYamlFiles
	manifest := c.String("manifest")
	if manifest != "" {
		paths = []string{manifest}
	}
	ctx, err := config.NewContext(paths...)
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

	ctx.Parameters = p
	ctx.Context = c.Context

	return ctx, nil
}
