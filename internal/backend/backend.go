package backend

import (
	"context"
)

type Backend interface {
	Store() Store
	Encryption() Encryption
}

func New(ctx context.Context, config BackendConfig) (Backend, error) {
	return NewAwsBackend(ctx, config)
}
