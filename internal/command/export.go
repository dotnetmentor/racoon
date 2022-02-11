package command

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/dotnetmentor/racoon/internal/aws"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/output"
	"github.com/dotnetmentor/racoon/internal/utils"

	"github.com/urfave/cli/v2"
)

func Export(ctx config.AppContext) *cli.Command {
	m := ctx.Manifest

	return &cli.Command{
		Name:  "export",
		Usage: "export secrets using outputs defined in the manifest file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "export a single output",
			},
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "export a single output to the specified path",
			},
			&cli.StringSliceFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage:   "include secret in export",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"e"},
				Usage:   "exclude secret from export",
			},
		},
		Action: func(c *cli.Context) error {
			ot := c.String("output")
			p := c.String("path")
			if ot == "" && p != "" {
				ctx.Log.Warn("the flag --path is not allowed without also specifying the --output flag")
				return nil
			}

			awsParameterStore, err := aws.NewParameterStoreClient(c.Context)
			if err != nil {
				return err
			}

			includes := c.StringSlice("include")
			excludes := c.StringSlice("exclude")
			context := c.String("context")

			// read from store
			secrets := []string{}
			values := map[string]string{}
			for _, s := range m.Secrets {
				if len(excludes) > 0 && utils.StringSliceContains(excludes, s.Name) {
					continue
				}
				if len(includes) > 0 && !utils.StringSliceContains(includes, s.Name) {
					continue
				}

				secrets = append(secrets, s.Name)

				if s.Default != nil {
					ctx.Log.Infof("reading %s from %s", s.Name, "default")
					values[s.Name] = *s.Default
				}

				if s.ValueFrom != nil {
					if s.ValueFrom.AwsParameterStore != nil {
						key := aws.ParameterStoreKey(m.Stores.AwsParameterStore, s, context)
						ctx.Log.Infof("reading %s from %s ( key=%s )", s.Name, config.StoreTypeAwsParameterStore, key)
						out, err := awsParameterStore.GetParameter(c.Context, &ssm.GetParameterInput{
							Name:           &key,
							WithDecryption: true,
						})
						if err != nil {
							return err
						}
						values[s.Name] = *out.Parameter.Value
					}
				}
			}

			// create outputs
			outputMatched := false
			for _, o := range m.Outputs {
				if ot != "" && string(o.Type) != ot {
					continue
				}

				outputMatched = true

				path := o.Path
				if p != "" {
					path = p
				}

				if ot == "" && path == "-" {
					ctx.Log.Infof("writing to stdout is only allowed when using the --output flag, skipping output %s", o.Type)
					continue
				}

				filtered := []string{}
				for _, s := range secrets {
					if len(o.Exclude) > 0 && utils.StringSliceContains(o.Exclude, s) {
						continue
					}
					if len(o.Include) > 0 && !utils.StringSliceContains(o.Include, s) {
						continue
					}
					filtered = append(filtered, s)
				}

				file := os.Stdout
				if path != "" && path != "-" {
					if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
						return fmt.Errorf("failed to open file for writing, %v", err)
					}
					defer file.Close()
					defer file.Sync()
				}
				w := bufio.NewWriter(file)
				defer w.Flush()

				switch o.Type {
				case config.OutputTypeDotenv:
					ctx.Log.Infof("exporting secrets as dotenv ( path=%s )", path)
					output.Dotenv(w, filtered, o.Map, values)
					break
				case config.OutputTypeTfvars:
					ctx.Log.Infof("exporting secrets as tfvars ( path=%s )", path)
					output.Tfvars(w, filtered, o.Map, values)
					break
				case config.OutputTypeJson:
					ctx.Log.Infof("exporting secrets as json ( path=%s )", path)
					output.Json(w, filtered, o.Map, values)
					break
				default:
					return fmt.Errorf("unsupported output type %s", o.Type)
				}
			}

			if ot != "" && !outputMatched {
				return fmt.Errorf("unknown output type %s", ot)
			}

			return nil
		},
	}
}
