package backend

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AwsS3BackendStore struct {
	Context   context.Context
	Config    AwsS3BackendConfig
	AwsConfig aws.Config
}

func (b AwsS3BackendStore) Upload(key string, body []byte) error {
	bucket := b.Config.Bucket
	bodyReader := bytes.NewReader(body)

	s3Client := s3.NewFromConfig(b.AwsConfig)
	_, err := s3Client.PutObject(b.Context, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bodyReader,
	})
	if err != nil {
		return err
	}
	return nil
}

func (b AwsS3BackendStore) Download(key string) ([]byte, error) {
	bucket := b.Config.Bucket

	s3Client := s3.NewFromConfig(b.AwsConfig)
	output, err := s3Client.GetObject(b.Context, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Body)
	return buf.Bytes(), nil
}

func (b AwsS3BackendStore) List() ([]string, error) {
	bucket := b.Config.Bucket

	s3Client := s3.NewFromConfig(b.AwsConfig)
	output, err := s3Client.ListObjectsV2(b.Context, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0)
	for _, o := range output.Contents {
		keys = append(keys, *o.Key)
	}
	return keys, nil
}
