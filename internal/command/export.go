package command

import (
	"fmt"
	"io/ioutil"

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
					ctx.Log.Debugf("reading %s from %s", s.Name, "default")
					values[s.Name] = *s.Default
				}

				if s.ValueFrom != nil {
					if s.ValueFrom.AwsParameterStore != nil {
						ctx.Log.Debugf("reading %s from %s", s.Name, config.StoreTypeAwsParameterStore)
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

				var out string
				switch o.Type {
				case config.OutputTypeDotenv:
					ctx.Log.Debugf("exporting secrets as dotenv (%s)", o.Path)
					out = output.Dotenv(m, values)
					ioutil.WriteFile(o.Path, []byte(out), 0600)
					break
				case config.OutputTypeTfvars:
					ctx.Log.Debugf("exporting secrets as tfvars (%s)", o.Path)
					out = output.Tfvars(m, values)
					ioutil.WriteFile(o.Path, []byte(out), 0600)
					break
				default:
					panic(fmt.Errorf("unsupported output type %s", o.Type))
				}
			}

			return nil
		},
	}
}
