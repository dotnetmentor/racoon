package aws

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/camelcase"

	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/utils"
)

func NewParameterStoreClient(ctx context.Context) (*ssm.Client, error) {
	if awsRegion := utils.StringEnvVar("AWS_REGION", ""); awsRegion == "" {
		return nil, fmt.Errorf("required environment variable AWS_REGION has no value set")
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return ssm.NewFromConfig(awsConfig), nil
}

func ParameterStoreKey(c config.AwsParameterStoreConfig, s config.SecretConfig, context string) string {
	key := s.ValueFrom.AwsParameterStore.Key
	if key == "" {
		key = c.DefaultKeyFormat
	}
	nameKey := camelCaseSplitToLowerJoinBySlashAndUnderscore(s.Name)
	key = strings.ReplaceAll(key, "{Context}", context)
	key = strings.ReplaceAll(key, "{Key}", nameKey)
	return key
}

func camelCaseSplitToLowerJoinBySlashAndUnderscore(name string) (key string) {
	parts := camelcase.Split(name)
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
	key = fmt.Sprintf("%s/%s", parts[0], strings.Join(parts[1:], "_"))
	return
}
