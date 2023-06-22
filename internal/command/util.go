package command

import (
	"fmt"
	"strings"

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

	p, err := parseParams(c.StringSlice("layer"))
	if err != nil {
		ctx.Log.Error(err)
		ctx.Log.Exit(1)
	}

	// validate parameters are defined by manifest
	for k := range p {
		if _, ok := ctx.Manifest.Config.Parameters[k]; !ok {
			return ctx, fmt.Errorf("parameter %s provided but not defined", k)
		}
	}

	// validate required parameters
	for k, v := range ctx.Manifest.Config.Parameters {
		if v.Required {
			if _, ok := p[k]; !ok {
				return ctx, fmt.Errorf("required parameter %s must be set", k)
			}
		}
	}

	ctx.Parameters = p
	ctx.Context = c.Context

	return ctx, nil
}

func parseParams(ls []string) (config.Parameters, error) {
	p := config.Parameters{}
	for _, l := range ls {
		parts := strings.Split(l, "=")
		// TODO: Error handling for layer flag value parsing
		lk := parts[0]
		lv := parts[1]
		p[lk] = lv
	}
	return p, nil
}

func replaceParams(s string, ls config.Parameters) string {
	for k, v := range ls {
		s = strings.ReplaceAll(s, fmt.Sprintf("{%s}", k), v)
	}
	return s
}
