package mongox

import (
	"context"

	"github.com/soyacen/gox/conc/lazyload"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"mongox",
	fx.Provide(NewClients),
	fx.Invoke(
		func(lc fx.Lifecycle, g *lazyload.Group[*mongo.Client]) {
			lc.Append(fx.StopHook(func(ctx context.Context) error {
				return g.Close(ctx)
			}))
		},
	),
)
