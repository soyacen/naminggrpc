package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scriptCmd = &cobra.Command{
	Use: "job",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("job called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)
}
