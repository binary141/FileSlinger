package cmd

import (
	"os"
	"strings"

	"github.com/binary141/FileSlinger/config"
	"github.com/binary141/FileSlinger/server"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server to receive files",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		dir, _ := cmd.Flags().GetString("dir")
		maxFiles, _ := cmd.Flags().GetInt("max-files")
		token, _ := cmd.Flags().GetString("token")
		relay, _ := cmd.Flags().GetString("relay")

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if !cmd.Flags().Changed("port") && cfg.Port != nil {
			port = *cfg.Port
		}
		if !cmd.Flags().Changed("dir") && cfg.Dir != nil {
			dir = *cfg.Dir
		}
		if !cmd.Flags().Changed("max-files") && cfg.MaxFiles != nil {
			maxFiles = *cfg.MaxFiles
		}
		if !cmd.Flags().Changed("token") && cfg.Token != nil {
			token = *cfg.Token
		}
		if !cmd.Flags().Changed("relay") && cfg.RelayURL != nil {
			relay = *cfg.RelayURL
		}

		if strings.HasPrefix(dir, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			dir = home + dir[1:]
		}

		serverCfg := server.Config{Port: port, Dir: dir, MaxFiles: maxFiles, Token: token, RelayURL: relay}
		if relay != "" {
			return server.StartRelay(serverCfg)
		}
		return server.Start(serverCfg)
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringP("dir", "d", "uploads", "Directory to save received files")
	serveCmd.Flags().IntP("max-files", "n", 0, "Max number of files to receive before shutting down (0 = unlimited)")
	serveCmd.Flags().StringP("token", "t", "", "Auth token (auto-generated if not set)")
	serveCmd.Flags().String("relay", "", "Relay server URL (enables cloud relay mode)")
}
