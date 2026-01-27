package cronjob

import "go.uber.org/fx"

var Module = fx.Module(
	"cronjob",
	fx.Provide(NewRepository, NewService),
)
