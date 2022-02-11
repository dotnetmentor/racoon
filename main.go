package main

import (
	"os"

	"github.com/dotnetmentor/racoon/internal/command"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx := config.NewContext()

	// configure cli app
	app := &cli.App{
		Name:  "racoon",
		Usage: "secrets are my thing",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"c"},
				Usage:   "sets the context",
				Value:   "local",
			},
		},
		Commands: []*cli.Command{
			command.Create(ctx),
			command.Export(ctx),
			command.Read(ctx),
		},
	}

	// run commands
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
