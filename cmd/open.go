package cmd

import "github.com/spf13/cobra"

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "The open commands opens a database connection.",
	Long:  `The open commands opens a database connection.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: open database
	},
}
