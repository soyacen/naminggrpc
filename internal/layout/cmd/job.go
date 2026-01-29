package cmd

import (
	"log/slog"

	"github.com/soyacen/grocer/grocer"
	"github.com/soyacen/grocer/internal/layout/config"
	"github.com/soyacen/grocer/internal/layout/internal/job/job"
	"github.com/soyacen/grocer/internal/layout/pkg/logx"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var jobCmd = &cobra.Command{
	Use: "job",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		app := fx.New(
			ContextModule(ctx),
			logx.Module(),
			config.Module,
			job.Module,
			grocer.Module,
			fx.WithLogger(func(slogger *slog.Logger) fxevent.Logger {
				fxlogger := &fxevent.SlogLogger{Logger: slogger}
				fxlogger.UseContext(ctx)
				return fxlogger
			}),
			fx.Invoke(
				func(lc fx.Lifecycle, s *job.Service) {
					lc.Append(fx.StartHook(s.Run))
				},
			),
		)
		return app.Start(ctx)
	},
}

func init() {
	rootCmd.AddCommand(jobCmd)
}
