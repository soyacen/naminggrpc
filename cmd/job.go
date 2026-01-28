package cmd

import (
	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "add job",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := flag.IsValid(); err != nil {
			return err
		}
		return jobRun(cmd, args, "job")
	},
}

func init() {
	rootCmd.AddCommand(jobCmd)
	jobCmd.Flags().StringVarP(&flag.Name, "name", "n", "", "job name, must consist of alphanumeric characters and underscores, and start with a letter, required")
	_ = jobCmd.MarkFlagRequired("name")
	jobCmd.Flags().StringVarP(&flag.Dir, "dir", "d", "", "project directory, default is current directory")
}
