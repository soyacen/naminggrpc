package grpcx

import (
	"context"
)

type Registrar interface {
	Register(ctx context.Context) error
	Deregister(ctx context.Context) error
}
