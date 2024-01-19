package store

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"

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
	pskf := sourceConfig.DefaultKey
	if len(propertySource.Key) > 0 {
		pskf = propertySource.Key
	}

	if len(pskf) == 0 {
		return api.NewValue(api.NewValueSource(layer, api.SourceTypeAwsParameterStore), "", "", missingKeyError(), sensitive || sourceConfig.ForceSensitive)
	}

	psk := awpParameterStoreKey(ctx.Parameters.Replace(pskf), key)
	ctx.Log.Debugf("reading %s from %s", psk, config.SourceTypeAwsParameterStore)
	out, err := s.client.GetParameter(ctx.Context, &ssm.GetParameterInput{
		Name:           &psk,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		var notFound *ssmtypes.ParameterNotFound
		if !errors.As(err, &notFound) {
			return api.NewValue(api.NewValueSource(layer, api.SourceTypeAwsParameterStore), psk, "", err, sensitive || sourceConfig.ForceSensitive)
		} else {
			treatAsError := sourceConfig.TreatNotFoundAsError
			if propertySource.TreatNotFoundAsError != nil {
				treatAsError = *propertySource.TreatNotFoundAsError
			}
			if treatAsError {
				ctx.Log.Warnf("%s not found in %s, configured to be treated as an error", psk, config.SourceTypeAwsParameterStore)
				return api.NewValue(api.NewValueSource(layer, api.SourceTypeAwsParameterStore), psk, "", fmt.Errorf("%s not found in %s, configured to be treated as an error, %s", psk, config.SourceTypeAwsParameterStore, notFound), sensitive || sourceConfig.ForceSensitive)
			}
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
		Overwrite:   aws.Bool(true),
	}

	if sourceConfig.KmsKey != "" {
		i.KeyId = &sourceConfig.KmsKey
	}

	if _, err := s.client.PutParameter(ctx.Context, &i); err != nil {
		ctx.Log.Errorf("failed to create parameter %s in %s, %v", key, config.SourceTypeAwsParameterStore, err)
		return err
	}

	tags := []ssmtypes.Tag{}

	if ctx.Manifest.Name != "" {
		tags = append(tags, ssmtypes.Tag{
			Key:   aws.String("racoon/owner"),
			Value: aws.String(ctx.Manifest.Name),
		})
	}

	tags = append(tags, ssmtypes.Tag{
		Key:   aws.String("racoon/version"),
		Value: aws.String(ctx.Metadata.Version),
	})

	for k, v := range ctx.Manifest.Labels {
		fv := ctx.Parameters.Replace(v)
		tags = append(tags, ssmtypes.Tag{
			Key:   aws.String(k),
			Value: aws.String(fv),
		})
	}

	if _, err := s.client.AddTagsToResource(ctx.Context, &ssm.AddTagsToResourceInput{
		ResourceId:   &key,
		ResourceType: ssmtypes.ResourceTypeForTaggingParameter,
		Tags:         tags,
	}); err != nil {
		ctx.Log.Errorf("failed to tag parameter %s in %s, %v", key, config.SourceTypeAwsParameterStore, err)
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
	nameKey := utils.FormatKey(key, utils.Formatting{
		Lowercase:     true,
		WordSeparator: "_",
		PathSeparator: "/",
	})
	key = strings.ReplaceAll(format, "{key}", nameKey)
	return key
}

// NOTE: Really ugly hack to avoid magic strings, poor performance expected
func missingKeyError() error {
	m := config.Manifest{}
	p := config.PropertyConfig{
		Source: &config.ValueSourceConfig{
			AwsParameterStore: &config.ValueFromAwsParameterStore{},
		},
	}
	configKey := strings.Join(tagsForFields(&m, &m.Config, &m.Config.Sources, &m.Config.Sources.AwsParameterStore, &m.Config.Sources.AwsParameterStore.DefaultKey), ".")
	sourceKey := strings.Join(tagsForFields(&p, &p.Source, &p.Source.AwsParameterStore, &p.Source.AwsParameterStore.Key), ".")
	return fmt.Errorf("key missing for %s, set %s or %s", api.SourceTypeAwsParameterStore, configKey, sourceKey)
}

func tagsForFields(fields ...interface{}) (tags []string) {
	for fi, f := range fields {
		if len(fields) > fi+1 {
			nfv := fields[fi+1]
			fv := reflect.ValueOf(f).Elem()
			if fv.Kind() == reflect.Ptr {
				fv = fv.Elem()
			}
			for i := 0; i < fv.NumField(); i++ {
				if fv.Field(i).Addr().Interface() == nfv {
					tag := fv.Type().Field(i).Tag.Get("yaml")
					tag = strings.ReplaceAll(tag, ",omitempty", "")
					tags = append(tags, tag)
					break
				}
			}
		}
	}
	return
}
