package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/dotnetmentor/racoon/internal/command"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const metadataExitCode string = "exitcode"

//go:embed ui/dist
var staticFiles embed.FS

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

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
		default:
			exitCode = 128
		}
	}

	os.Exit(exitCode)
}

func createApp() (*cli.App, config.AppContext) {
	metadata := config.AppMetadata{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	ctx, err := config.NewContext(metadata)
	if err != nil {
		ctx.Log.Error(err)
		ctx.Log.Exit(1)
	}

	// configure cli app
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s %s\ncommit = %s\ndate = %s\n", c.App.Name, c.App.Version, metadata.Commit, metadata.Date)
	}
	app := &cli.App{
		Name:    "racoon",
		Usage:   "configuration and secrets management",
		Version: metadata.Version,
		CommandNotFound: func(c *cli.Context, s string) {
			ctx.Log.Warnf("unknown command %s", s)
			c.App.Metadata[metadataExitCode] = 127
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "manifest",
				Aliases: []string{"m"},
				Usage:   "path to the manifest file",
				Value:   "",
			},
			&cli.StringFlag{
				Name:    "loglevel",
				Aliases: []string{"l"},
				Usage:   "sets the log level",
				Value:   "info",
			},
		},
		Commands: []*cli.Command{
			command.Export(metadata),
			command.Read(metadata),
			command.Write(metadata),
			command.Config(metadata),
			command.UI(metadata, staticFiles),
		},
		Before: func(c *cli.Context) error {
			l := c.String("loglevel")
			level, err := logrus.ParseLevel(l)
			if err != nil {
				return err
			}
			ctx.Log.SetLevel(level)
			return nil
		},
	}

	logo := `
 ______     ______     ______     ______     ______     __   __
/\  == \   /\  __ \   /\  ___\   /\  __ \   /\  __ \   /\ "-.\ \
\ \  __<   \ \  __ \  \ \ \____  \ \ \/\ \  \ \ \/\ \  \ \ \-.  \
 \ \_\ \_\  \ \_\ \_\  \ \_____\  \ \_____\  \ \_____\  \ \_\\"\_\
  \/_/ /_/   \/_/\/_/   \/_____/   \/_____/   \/_____/   \/_/ \/_/`

	url := "https://github.com/dotnetmentor/racoon"

	app.CustomAppHelpTemplate = fmt.Sprintf(`
%s
%66s

%s
`, logo, url, cli.AppHelpTemplate)

	return app, ctx
}
