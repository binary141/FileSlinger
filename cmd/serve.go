package cmd

import (
	"fileSlinger/server"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server to receive files",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		dir, _ := cmd.Flags().GetString("dir")
		return server.Start(server.Config{Port: port, Dir: dir})
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringP("dir", "d", "uploads", "Directory to save received files")
}
