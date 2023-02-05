package command

import (
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/urfave/cli/v2"
)

func getContext(c *cli.Context) config.AppContext {
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
	return ctx
}
