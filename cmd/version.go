package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v0.0.25"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version number of grocer",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version: ", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
