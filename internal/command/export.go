package command

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/dotnetmentor/racoon/internal/aws"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/output"

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
			},
		},
		Action: func(c *cli.Context) error {
			awsParameterStore, err := aws.NewParameterStoreClient(c.Context)
			if err != nil {
				return err
			}

			// read from store
			values := map[string]string{}
			for _, s := range m.Secrets {
				if s.Default != nil {
					ctx.Log.Infof("reading %s from %s", s.Name, "default")
					values[s.Name] = *s.Default
				}

				if s.ValueFrom != nil {
					if s.ValueFrom.AwsParameterStore != nil {
						ctx.Log.Infof("reading %s from %s", s.Name, config.StoreTypeAwsParameterStore)
						out, err := awsParameterStore.GetParameter(c.Context, &ssm.GetParameterInput{
							Name:           &s.ValueFrom.AwsParameterStore.Key,
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
			for _, o := range m.Outputs {
				ot := c.String("output")
				if ot != "" && string(o.Type) != ot {
					continue
				}

				if ot == "" && o.Path == "-" {
					ctx.Log.Infof("writing to stdout is only allowed when using the --output flag, skipping output %s", o.Type)
					continue
				}

				file := os.Stdout
				if o.Path != "-" {
					if file, err = os.OpenFile(o.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
						return fmt.Errorf("failed to open file for writing, %v", err)
					}
					defer file.Close()
					defer file.Sync()
				}
				w := bufio.NewWriter(file)
				defer w.Flush()

				switch o.Type {
				case config.OutputTypeDotenv:
					ctx.Log.Infof("exporting secrets as dotenv ( path=%s )", o.Path)
					output.Dotenv(w, m, values)
					break
				case config.OutputTypeTfvars:
					ctx.Log.Infof("exporting secrets as tfvars ( path=%s )", o.Path)
					output.Tfvars(w, m, values)
					break
				default:
					panic(fmt.Errorf("unsupported output type %s", o.Type))
				}
			}

			return nil
		},
	}
}
