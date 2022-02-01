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
		Usage: "working with secrets is my thing",
		Commands: []*cli.Command{
			command.Export(ctx),
		},
	}

	// run commands
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
