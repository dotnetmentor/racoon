package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/fatih/camelcase"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/utils"
)

func newAwsParameterStore(ctx context.Context) (*AwsParameterStore, error) {
	client, err := newParameterStoreClient(ctx)
	return &AwsParameterStore{
		client: client,
	}, err
}

type AwsParameterStore struct {
	client *ssm.Client
}

func (s *AwsParameterStore) Read(ctx config.AppContext, layer api.Layer, key string, sensitive bool, propertySource config.ValueFromAwsParameterStore, sourceConfig config.AwsParameterStoreConfig) api.Value {
	pskf := sourceConfig.KeyFormat
	if len(propertySource.Key) > 0 {
		pskf = propertySource.Key
	}

	psk := awpParameterStoreKey(config.ReplaceParams(pskf, ctx.Parameters), key)
	ctx.Log.Debugf("reading %s from %s", psk, config.SourceTypeAwsParameterStore)
	out, err := s.client.GetParameter(ctx.Context, &ssm.GetParameterInput{
		Name:           &psk,
		WithDecryption: true,
	})
	if err != nil {
		var notFound *ssmtypes.ParameterNotFound
		if !errors.As(err, &notFound) {
			return api.NewValue(api.NewValueSource(layer, api.SourceTypeAwsParameterStore), psk, "", err, sensitive || sourceConfig.ForceSensitive)
		} else {
			ctx.Log.Debugf("%s not found in %s", psk, config.SourceTypeAwsParameterStore)
			return api.NewValue(api.NewValueSource(layer, api.SourceTypeAwsParameterStore), psk, "", api.NewNotFoundError(notFound, psk, api.SourceTypeAwsParameterStore), sensitive || sourceConfig.ForceSensitive)
		}
	} else {
		return api.NewValue(api.NewValueSource(layer, api.SourceTypeAwsParameterStore), psk, *out.Parameter.Value, err, sensitive || sourceConfig.ForceSensitive)
	}
}

func (s *AwsParameterStore) Write(ctx config.AppContext, key, value, description string, sourceConfig config.AwsParameterStoreConfig) error {
	ctx.Log.Infof("upserting parameter %s in %s", key, api.SourceTypeAwsParameterStore)
	i := ssm.PutParameterInput{
		Name:        &key,
		Description: &description,
		Value:       &value,
		Type:        ssmtypes.ParameterTypeSecureString,
		Tier:        ssmtypes.ParameterTierStandard,
		Overwrite:   true,
	}
	if sourceConfig.KmsKey != "" {
		i.KeyId = &sourceConfig.KmsKey
	}
	_, err := s.client.PutParameter(ctx.Context, &i)
	if err != nil {
		ctx.Log.Errorf("failed to create parameter %s in %s, %v", key, config.SourceTypeAwsParameterStore, err)
		return err
	}
	return nil
}

func newParameterStoreClient(ctx context.Context) (*ssm.Client, error) {
	if awsRegion := utils.StringEnvVar("AWS_REGION", ""); awsRegion == "" {
		return nil, fmt.Errorf("required environment variable AWS_REGION has no value set")
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return ssm.NewFromConfig(awsConfig), nil
}

func awpParameterStoreKey(format, key string) string {
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
