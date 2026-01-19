package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "A brief description of your command",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("script called")
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scriptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scriptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
