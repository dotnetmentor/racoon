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

			context := c.String("context")
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
						key := aws.ParameterStoreKey(m.Stores.AwsParameterStore, s, context)
						ctx.Log.Debugf("reading %s from %s ( key=%s )", s.Name, config.StoreTypeAwsParameterStore, key)
						out, err := awsParameterStore.GetParameter(c.Context, &ssm.GetParameterInput{
							Name:           &key,
							WithDecryption: true,
						})
						if err != nil {
							return err
						}
						value = *out.Parameter.Value
					}
				}

				fmt.Printf("%s", value)
				return nil
			}

			return fmt.Errorf("secret matching name %s was not found", name)
		},
	}
}
