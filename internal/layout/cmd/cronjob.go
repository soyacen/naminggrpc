package cmd

import (
	"os"
	"strconv"
	"time"

	"github.com/soyacen/gonfig/resource/nacos"
	"github.com/soyacen/gox/errorx"
	"github.com/soyacen/grocer/grocer"
	"github.com/soyacen/grocer/grocer/nacosx"
	"github.com/soyacen/grocer/internal/layout/config"
	"github.com/soyacen/grocer/internal/layout/internal/cronjob"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var cronjobCmd = &cobra.Command{
	Use:   "cronjob",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		configClient, err := nacosx.NewConfigClient(&nacosx.Options{
			IpAddr:    wrapperspb.String(os.Getenv("NACOS_CONFIG_IP")),
			Port:      wrapperspb.UInt64(errorx.Ignore(strconv.ParseUint(os.Getenv("NACOS_CONFIG_PORT"), 10, 64))),
			Namespace: wrapperspb.String(os.Getenv("NACOS_CONFIG_NAMESPACE")),
			TimeoutMs: durationpb.New(5 * time.Second),
		})
		if err != nil {
			return err
		}
		resource, err := nacos.New(configClient, os.Getenv("NACOS_CONFIG_GROUP"), os.Getenv("NACOS_CONFIG_DATA_ID"))
		if err != nil {
			return err
		}
		if err := config.LoadConfig(ctx, resource); err != nil {
			return err
		}
		app := fx.New(
			ContextModule(ctx),
			config.Module,
			cronjob.Module,
			grocer.Module,
		)
		return app.Start(ctx)
	},
}

func init() {
	rootCmd.AddCommand(cronjobCmd)
}
