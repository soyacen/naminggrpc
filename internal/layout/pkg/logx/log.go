package logx

import (
	"log/slog"
	"os"

	"github.com/soyacen/gox/slogx"
	"github.com/soyacen/grocer/internal/layout/config"
	"go.uber.org/fx"
)

var Level *slog.LevelVar

func Init() {
	Level = &slog.LevelVar{}
	if config.IsDev() {
		Level.Set(slog.LevelDebug)
	} else {
		Level.Set(slog.LevelInfo)
	}
	slog.SetDefault(slog.New(slogx.WithLevel(slogx.WithContext(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})), Level)))
}

func Module() fx.Option {
	return fx.Module("slog",
		fx.Provide(func() *slog.Logger { return slog.Default() }),
	)
}
