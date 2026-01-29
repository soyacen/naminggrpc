package http

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"http",
	fx.Provide(NewRepository, NewService),
)
