package command

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/dotnetmentor/racoon/internal/aws"
	"github.com/dotnetmentor/racoon/internal/config"

	"github.com/urfave/cli/v2"
)

func Read(ctx config.AppContext) *cli.Command {
	m := ctx.Manifest

	return &cli.Command{
		Name:  "read",
		Usage: "reads a single secret value from it's store",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			awsParameterStore, err := aws.NewParameterStoreClient(c.Context)
			if err != nil {
				return err
			}

			name := strings.ToLower(strings.TrimSpace(c.Args().First()))

			// read from store
			for _, s := range m.Secrets {
				if strings.ToLower(s.Name) != name {
					continue
				}

				value := ""
				if s.Default != nil {
					ctx.Log.Infof("reading %s from %s", s.Name, "default")
					value = *s.Default
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
						value = *out.Parameter.Value
					}

					fmt.Printf("%s", value)
				}
			}

			return fmt.Errorf("secret matching name %s was not found", name)
		},
	}
}
