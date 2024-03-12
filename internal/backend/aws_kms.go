package backend

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

type AwsKmsEncryption struct {
	Context   context.Context
	Config    AwsKmsBackendConfig
	AwsConfig aws.Config
}

func (encryption AwsKmsEncryption) Encrypt(v []byte) ([]byte, error) {
	// Key
	keyId := encryption.Config.KmsKey
	if len(keyId) == 0 {
		return nil, fmt.Errorf("kms key not set")
	}

	// Create KMS service client
	kmsClient := kms.NewFromConfig(encryption.AwsConfig)

	// Encrypt the data
	er, err := kmsClient.Encrypt(encryption.Context, &kms.EncryptInput{
		KeyId:     aws.String(keyId),
		Plaintext: v,
	})

	if err != nil {
		fmt.Println("Got error encrypting data: ", err)
		return nil, err
	}

	return []byte(base64.StdEncoding.EncodeToString(er.CiphertextBlob)), nil
}

func (encryption AwsKmsEncryption) Decrypt(v []byte) ([]byte, error) {
	keyId := encryption.Config.KmsKey
	if len(keyId) == 0 {
		return nil, fmt.Errorf("kms key not set")
	}

	bytes, err := base64.StdEncoding.DecodeString(string(v))
	if err != nil {
		return nil, err
	}

	// Create KMS service client
	svc := kms.NewFromConfig(encryption.AwsConfig)

	// Decrypt the data
	er, err := svc.Decrypt(encryption.Context, &kms.DecryptInput{
		KeyId:          aws.String(keyId),
		CiphertextBlob: bytes,
	})

	if err != nil {
		fmt.Println("Got error decrypting data: ", err)
		return nil, err
	}

	return er.Plaintext, nil
}
