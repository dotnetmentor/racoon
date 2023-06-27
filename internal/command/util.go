package command

import (
	"fmt"

	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func getContext(c *cli.Context) (config.AppContext, error) {
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

	// validate parameters are defined by manifest
	for k := range p {
		if _, ok := ctx.Manifest.Config.Parameters[k]; !ok {
			return ctx, fmt.Errorf("parameter %s, provided but not defined", k)
		}
	}

	// validate required parameters
	for k, v := range ctx.Manifest.Config.Parameters {
		if v.Required {
			if _, ok := p[k]; !ok {
				return ctx, fmt.Errorf("required parameter must be set, parameter: %s", k)
			}
		}
	}

	ctx.Parameters = p
	ctx.Context = c.Context

	return ctx, nil
}
