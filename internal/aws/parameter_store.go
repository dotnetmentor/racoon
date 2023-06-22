package aws

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/camelcase"

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

func ParameterStoreKey(format, key string) string {
	nameKey := camelCaseSplitToLowerJoinBySlashAndUnderscore(key)
	key = strings.ReplaceAll(format, "{key}", nameKey)
	return key
}

func camelCaseSplitToLowerJoinBySlashAndUnderscore(name string) (key string) {
	parts := camelcase.Split(name)
	if len(parts) == 1 {
		return parts[0]
	}

	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
	return fmt.Sprintf("%s/%s", parts[0], strings.Join(parts[1:], "_"))
}
