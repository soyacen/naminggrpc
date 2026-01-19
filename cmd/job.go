package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("script called")
	},
}

func init() {
	rootCmd.AddCommand(jobCmd)
}
