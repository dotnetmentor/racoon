package backend

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type Encryption interface {
	Encrypt(val []byte) ([]byte, error)
	Decrypt(val []byte) ([]byte, error)
}

func NewEncryption(ctx context.Context, config EncryptionConfig, awsConfig aws.Config) (Encryption, error) {
	if config.AwsKms != nil {
		encryption := &AwsKmsEncryption{
			Context:   ctx,
			Config:    *config.AwsKms,
			AwsConfig: awsConfig,
		}
		return encryption, nil
	} else {
		return nil, fmt.Errorf("encryption not configured")
	}
}
