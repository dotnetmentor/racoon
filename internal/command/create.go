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

func Create() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "create missing secrets defined in the manifest file",
		Action: func(c *cli.Context) error {
			ctx, err := getContext(c)
			if err != nil {
				return err
			}
			m := ctx.Manifest

			promptForValue := func(p config.PropertyConfig) string {
				fmt.Printf("%s? %s%s (%s) ", chalk.Green, chalk.White, p.Name, p.Description)
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

			awsParameterStore, err := aws.NewParameterStoreClient(c.Context)
			if err != nil {
				return err
			}

			// create missing in store
			for _, p := range m.Properties {
				if p.Source != nil {
					if p.Source.AwsParameterStore != nil {
						pskf := m.Config.Sources.AwsParameterStore.KeyFormat
						if len(p.Source.AwsParameterStore.Key) > 0 {
							pskf = p.Source.AwsParameterStore.Key
						}
						key := aws.ParameterStoreKey(replaceParams(pskf, ctx.Parameters), p.Name)

						ctx.Log.Infof("checking if %s exists in %s ( key=%s )", p.Name, config.SourceTypeAwsParameterStore, key)
						_, err := awsParameterStore.GetParameter(c.Context, &ssm.GetParameterInput{
							Name:           &key,
							WithDecryption: true,
						})
						if err != nil {
							var notFound *types.ParameterNotFound
							if errors.As(err, &notFound) {
								value := promptForValue(p)
								hasValue := len(value) > 0
								if !hasValue && p.Default != nil {
									if promptYesNo(fmt.Sprintf("no value was provided for secret %s, continue", p.Name)) {
										continue
									}
								}
								if hasValue {
									ctx.Log.Infof("creating parameter %s in %s", key, config.SourceTypeAwsParameterStore)
									i := ssm.PutParameterInput{
										Name:        &key,
										Description: &p.Description,
										Value:       &value,
										Type:        types.ParameterTypeSecureString,
										Tier:        types.ParameterTierStandard,
									}
									if m.Config.Sources.AwsParameterStore.KmsKey != "" {
										i.KeyId = &m.Config.Sources.AwsParameterStore.KmsKey
									}
									_, err := awsParameterStore.PutParameter(c.Context, &i)
									if err != nil {
										ctx.Log.Errorf("failed to create parameter %s in %s, %v", key, config.SourceTypeAwsParameterStore, err)
										return err
									}
									continue
								} else {
									return fmt.Errorf("no value was provided for secret %s", p.Name)
								}
							} else {
								ctx.Log.Errorf("failed to get parameter %s from %s, %v", key, config.SourceTypeAwsParameterStore, err)
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
