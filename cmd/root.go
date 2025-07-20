package cmd

import (
	"os"

	"github.com/gkits/pavosql/pkg/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "pavosql",
	Short:   "pavosql is a simple, lightweight and single file based relational database.",
	Long:    `pavosql is a simple, lightweight and single file based relational database.`,
	Version: version.Version(),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(openCmd)
}
