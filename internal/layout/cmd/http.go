package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use: "http",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("http called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)
}
