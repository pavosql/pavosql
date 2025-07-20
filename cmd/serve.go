package cmd

import (
	"github.com/spf13/cobra"
)

var (
	port     uint16
	filePath string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "The serve command starts a database server.",
	Long:  `The serve command starts a database server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: start server
	},
}

func init() {
	serveCmd.Flags().
		StringVarP(&filePath, "file", "f", "/var/lib/pavosql/pavosql.db", "The database file that is served.")
	serveCmd.Flags().Uint16VarP(&port, "port", "p", 5225, "The port the database server is served on.")
}
