package cronjob

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"cronjob",
	fx.Provide(NewRepository, NewService),
	fx.Invoke(RunCronjob),
)

func RunCronjob(lc fx.Lifecycle, s *Service) {
	lc.Append(fx.StartHook(s.Run))
}
