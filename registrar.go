package naminggrpc

import (
	"context"
)

type Registrar interface {
	Register(ctx context.Context) error
	Deregister(ctx context.Context) error
}

type Factory interface {
	New(ctx context.Context, dsn string) (Registrar, error)
}
