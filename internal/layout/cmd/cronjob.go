package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cronjobCmd = &cobra.Command{
	Use:   "cronjob",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cronjob called")
	},
}

func init() {
	rootCmd.AddCommand(cronjobCmd)
}
