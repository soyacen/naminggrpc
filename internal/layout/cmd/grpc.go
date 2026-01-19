package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var grpcCmd = &cobra.Command{
	Use: "grpc",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("grpc called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)
}
