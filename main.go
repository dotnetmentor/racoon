package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/arsham/figurine/figurine"
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
		Name:            "racoon",
		Usage:           "configuration and secrets management",
		Version:         metadata.Version,
		HideHelpCommand: true,
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

	banner := newBanner("racoon", "https://github.com/dotnetmentor/racoon")

	cli.AppHelpTemplate = banner + cli.AppHelpTemplate
	cli.CommandHelpTemplate = banner + cli.CommandHelpTemplate
	cli.SubcommandHelpTemplate = banner + cli.SubcommandHelpTemplate

	for _, c := range app.Commands {
		c.HideHelpCommand = true
	}

	return app, ctx
}

func newBanner(text, url string) string {
	term := os.Getenv("TERM")
	nocolor := os.Getenv("NO_COLOR")
	if nocolor == "true" || term == "" || term == "dumb" {
		return fmt.Sprintf("\n%s - %s\n\n", strings.ToUpper(text), url)
	}

	logo := rainbow(text)
	return fmt.Sprintf("\n%s\n%66s\n\n", logo, url)
}

func rainbow(s string) string {
	var buf bytes.Buffer
	if err := figurine.Write(&buf, s, "Sub-Zero.flf"); err != nil {
		return err.Error()
	}
	return buf.String()
}
