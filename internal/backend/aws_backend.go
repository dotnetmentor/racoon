package backend

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

type AwsBackend struct {
	store      Store
	encryption Encryption
}

func NewAwsBackend(ctx context.Context, config BackendConfig) (Backend, error) {
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	store, err := NewStore(ctx, config.Store, awsConfig)
	if err != nil {
		return nil, err
	}

	encryption, err := NewEncryption(ctx, config.Encryption, awsConfig)
	if err != nil {
		return nil, err
	}

	return &AwsBackend{
		store:      store,
		encryption: encryption,
	}, nil
}

func (b *AwsBackend) Store() Store {
	return b.store
}

func (b *AwsBackend) Encryption() Encryption {
	return b.encryption
}
