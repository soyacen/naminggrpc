package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("script called")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
