package redisx

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/soyacen/gox/conc/lazyload"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"mongox",
	fx.Provide(NewClients),
	fx.Invoke(
		func(lc fx.Lifecycle, g *lazyload.Group[redis.UniversalClient]) {
			lc.Append(fx.StopHook(func(ctx context.Context) error {
				return g.Close(ctx)
			}))
		},
	),
)
