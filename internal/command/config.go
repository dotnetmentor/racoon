package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func Config(metadata config.AppMetadata) *cli.Command {
	return &cli.Command{
		Name:      "config",
		Usage:     "Manages configuration",
		UsageText: "",
		Hidden:    false,
		Flags:     []cli.Flag{},
		Subcommands: []*cli.Command{
			{
				Name:      "init",
				Usage:     "Generates a new configuration",
				UsageText: "",
				Action: func(c *cli.Context) error {
					exampleName := "my-service"
					examplePort := "1337"
					exampleParameter := "context"

					basepath, _ := os.Getwd()
					if len(basepath) > 0 {
						exampleName = filepath.Base(basepath)
					}

					m := config.Manifest{
						MetadataConfig: config.MetadataConfig{
							Name: exampleName,
						},
						Config: config.Config{
							Parameters: config.ParameterConfigList{
								{
									Key:      exampleParameter,
									Required: true,
								},
							},
						},
						Properties: config.PropertyList{
							{
								Name:        "Environment",
								Description: "The port to listen on",
								Sensitive:   false,
								Source: &config.ValueSourceConfig{
									Parameter: &exampleParameter,
								},
							},
							{
								Name:        "Http.Port",
								Description: "The port to listen on",
								Sensitive:   false,
								Default:     &examplePort,
							},
						},
						Outputs: config.OutputList{
							{
								Type: config.OutputTypeJson,
								Paths: []string{
									"./config.json",
								},
							},
						},
					}

					b, err := yaml.Marshal(m)
					if err != nil {
						return err
					}

					path := manifestPaths(c)[0]

					c.App.Writer.Write([]byte(fmt.Sprintf("generating manifest file %s\n", path)))
					if !filepath.IsAbs(path) {
						path = filepath.Join(basepath, path)
					}

					if _, err := os.Stat(path); !os.IsNotExist(err) {
						return fmt.Errorf("manifest file already exists, %s", path)
					}

					file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						return fmt.Errorf("failed to open file for writing, %v", err)
					}
					defer file.Close()
					defer file.Sync()

					if _, err := file.Write(b); err != nil {
						return fmt.Errorf("failed to write manifest file, %v", err)
					}

					return nil
				},
			},
			{
				Name:      "show",
				Usage:     "Shows the current configuration",
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
