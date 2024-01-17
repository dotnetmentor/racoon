package command

import (
	"fmt"

	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func Config(metadata config.AppMetadata) *cli.Command {
	return &cli.Command{
		Name:      "config",
		Usage:     "Manage configuration",
		UsageText: "",
		Hidden:    false,
		Flags:     []cli.Flag{},
		Subcommands: []*cli.Command{
			{
				Name:      "show",
				Usage:     "Show configuration",
				UsageText: "",
				Action: func(c *cli.Context) error {
					ctx, err := newContext(c, metadata, false)
					if err != nil {
						return err
					}

					b, err := yaml.Marshal(ctx.Manifest)
					if err != nil {
						return err
					}

					fmt.Println(string(b))

					return nil
				},
			},
		},
	}
}
