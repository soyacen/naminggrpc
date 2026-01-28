package dbx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/soyacen/gox/conc/lazyload"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"dbx",
	fx.Provide(NewDBs, NewSqlxDBs),
	fx.Invoke(
		func(lc fx.Lifecycle, g *lazyload.Group[*sql.DB]) {
			lc.Append(fx.StopHook(func(ctx context.Context) error {
				return g.Close(ctx)
			}))
		},
	),
	fx.Invoke(
		func(lc fx.Lifecycle, g *lazyload.Group[*sqlx.DB]) {
			lc.Append(fx.StopHook(func(ctx context.Context) error {
				return g.Close(ctx)
			}))
		},
	),
)
