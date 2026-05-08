package cmd

import (
	"fmt"

	"github.com/binary141/FileSlinger/client"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [files...]",
	Short: "Send one or more files to a fileslinger server",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		token, _ := cmd.Flags().GetString("token")
		relay, _ := cmd.Flags().GetString("relay")

		if relay != "" {
			if token == "" {
				return fmt.Errorf("--token is required")
			}
			return client.SendFilesRelay(relay, token, args)
		}

		if host == "" {
			return fmt.Errorf("--host is required (or use --relay for cloud relay mode)")
		}
		if token == "" {
			return fmt.Errorf("--token is required")
		}
		return client.SendFiles(host, port, token, args)
	},
}

func init() {
	sendCmd.Flags().StringP("host", "H", "", "Server host or IP address")
	sendCmd.Flags().IntP("port", "p", 8080, "Server port")
	sendCmd.Flags().StringP("token", "t", "", "Auth token (required)")
	sendCmd.Flags().String("relay", "", "Relay server URL (send via cloud relay instead of direct connection)")
}
