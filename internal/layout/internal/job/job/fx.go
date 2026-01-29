package job

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"job",
	fx.Provide(NewRepository, NewService),
)
