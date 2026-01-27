package cmd

import (
	"os"

	"github.com/soyacen/grocer/internal/layout/config"
	"github.com/soyacen/grocer/internal/layout/pkg/logx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "grocer",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logx.Init()
		if err := config.LoadConfigFromNacos(ctx); err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
