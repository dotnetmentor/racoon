package main

import (
	"os"

	"github.com/dotnetmentor/racoon/internal/command"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/urfave/cli/v2"
)

const metadataExitCode string = "exitcode"

func main() {
	app, ctx := createApp()

	// run commands
	exitCode := 0
	err := app.Run(os.Args)
	if err != nil {
		ctx.Log.Error(err)
		exitCode = 1
	}

	if app.Metadata[metadataExitCode] != nil {
		switch metaExitCode := app.Metadata[metadataExitCode].(type) {
		case int:
			exitCode = metaExitCode
			break
		default:
			exitCode = 128
			break
		}
	}

	os.Exit(exitCode)
}

func createApp() (*cli.App, config.AppContext) {
	ctx, err := config.NewContext()
	if err != nil {
		ctx.Log.Error(err)
		ctx.Log.Exit(1)
	}

	// configure cli app
	app := &cli.App{
		Name:  "racoon",
		Usage: "secrets are my thing",
		CommandNotFound: func(c *cli.Context, s string) {
			ctx.Log.Warnf("unknown command %s", s)
			c.App.Metadata[metadataExitCode] = 127
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"c"},
				Usage:   "sets the context",
				Value:   "default",
			},
			&cli.StringFlag{
				Name:    "manifest",
				Aliases: []string{"m"},
				Usage:   "path to manifest manifest file",
				Value:   "",
			},
		},
		Commands: []*cli.Command{
			command.Create(),
			command.Export(),
			command.Read(),
		},
	}
	return app, ctx
}
