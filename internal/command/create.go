package command

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/dotnetmentor/racoon/internal/aws"
	"github.com/dotnetmentor/racoon/internal/config"

	"github.com/urfave/cli/v2"
)

func Create(ctx config.AppContext) *cli.Command {
	m := ctx.Manifest

	return &cli.Command{
		Name:  "create",
		Usage: "create missing secrets defined in the manifest file",
		Action: func(c *cli.Context) error {
			awsParameterStore, err := aws.NewParameterStoreClient(c.Context)
			if err != nil {
				return err
			}

			// create missing in store
			for _, s := range m.Secrets {
				if s.ValueFrom != nil {
					if s.ValueFrom.AwsParameterStore != nil {
						ctx.Log.Infof("reading %s from %s", s.Name, "awsParameterStore")
						_, err := awsParameterStore.GetParameter(c.Context, &ssm.GetParameterInput{
							Name:           &s.ValueFrom.AwsParameterStore.Key,
							WithDecryption: true,
						})
						if err != nil {
							var notFound *types.ParameterNotFound
							if errors.As(err, &notFound) {
								ctx.Log.Infof("TODO create this parameter")
								continue
							} else {
								ctx.Log.Errorf("failed to get parameter %s, %v", s.ValueFrom.AwsParameterStore.Key, err)
								return err
							}
						} else {
							continue
						}
					}
				}
			}

			return nil
		},
	}
}
