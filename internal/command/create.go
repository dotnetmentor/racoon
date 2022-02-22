package command

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ttacon/chalk"

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
			promptForValue := func(s config.SecretConfig) string {
				fmt.Printf("%s? %s%s (%s) ", chalk.Green, chalk.White, s.Name, s.Description)
				reader := bufio.NewReader(os.Stdin)
				value, _ := reader.ReadString('\n')
				value = strings.TrimSuffix(value, "\n")
				return value
			}

			promptYesNo := func(msg string) bool {
				fmt.Printf("%s? %s%s (yes/no) ", chalk.Green, chalk.White, msg)
				reader := bufio.NewReader(os.Stdin)
				value, _ := reader.ReadString('\n')
				value = strings.TrimSuffix(value, "\n")
				if value == "yes" || value == "y" {
					return true
				}
				return false
			}

			context := c.String("context")

			awsParameterStore, err := aws.NewParameterStoreClient(c.Context)
			if err != nil {
				return err
			}

			// create missing in store
			for _, s := range m.Secrets {
				if s.ValueFrom != nil {
					if s.ValueFrom.AwsParameterStore != nil {
						key := aws.ParameterStoreKey(m.Stores.AwsParameterStore, s, context)

						ctx.Log.Infof("checking if %s exists in %s ( key=%s )", s.Name, config.StoreTypeAwsParameterStore, key)
						_, err := awsParameterStore.GetParameter(c.Context, &ssm.GetParameterInput{
							Name:           &key,
							WithDecryption: true,
						})
						if err != nil {
							var notFound *types.ParameterNotFound
							if errors.As(err, &notFound) {
								value := promptForValue(s)
								hasValue := len(value) > 0
								if !hasValue && s.Default != nil {
									if promptYesNo(fmt.Sprintf("no value was provided for secret %s, continue", s.Name)) {
										continue
									}
								}
								if hasValue {
									ctx.Log.Infof("creating parameter %s in %s", key, config.StoreTypeAwsParameterStore)
									i := ssm.PutParameterInput{
										Name:        &key,
										Description: &s.Description,
										Value:       &value,
										Type:        types.ParameterTypeSecureString,
										Tier:        types.ParameterTierStandard,
									}
									if m.Stores.AwsParameterStore.KmsKey != "" {
										i.KeyId = &m.Stores.AwsParameterStore.KmsKey
									}
									_, err := awsParameterStore.PutParameter(c.Context, &i)
									if err != nil {
										ctx.Log.Errorf("failed to create parameter %s in %s, %v", key, config.StoreTypeAwsParameterStore, err)
										return err
									}
									continue
								} else {
									return fmt.Errorf("no value was provided for secret %s", s.Name)
								}
							} else {
								ctx.Log.Errorf("failed to get parameter %s from %s, %v", key, config.StoreTypeAwsParameterStore, err)
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
