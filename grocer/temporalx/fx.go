package temporalx

import (
	"context"

	"github.com/soyacen/gox/conc/lazyload"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"temporalx",
	fx.Provide(NewClients),
	fx.Invoke(
		func(lc fx.Lifecycle, g *lazyload.Group[client.Client]) {
			lc.Append(fx.StopHook(func(ctx context.Context) error {
				return g.Close(ctx)
			}))
		},
	),
)
