package backend

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type Store interface {
	Upload(key string, body []byte) error
	Download(key string) ([]byte, error)
	List() ([]string, error)
}

func NewStore(ctx context.Context, config StoreConfig, awsConfig aws.Config) (Store, error) {
	if config.AwsS3 != nil {
		store := &AwsS3BackendStore{
			Context:   ctx,
			Config:    *config.AwsS3,
			AwsConfig: awsConfig,
		}
		return store, nil
	} else {
		return nil, fmt.Errorf("no backend configured")
	}
}
