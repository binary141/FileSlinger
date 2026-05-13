package cmd

import (
	"fmt"
	"strings"

	"github.com/binary141/fileslinger/client"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [files...]",
	Short: "Send one or more files to a fileslinger server",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		useHTTPS, _ := cmd.Flags().GetBool("https")
		token, _ := cmd.Flags().GetString("token")
		relay, _ := cmd.Flags().GetString("relay")

		if token == "" {
			return fmt.Errorf("--token is required")
		}

		if host == "" && relay == "" {
			return fmt.Errorf("--host is required (or use --relay for cloud relay mode)")
		}

		var uploadURL string
		if relay != "" {
			uploadURL = strings.TrimRight(relay, "/") + "/upload/" + token
		} else {
			scheme := "http"
			if useHTTPS {
				scheme = "https"
			}
			if !cmd.Flags().Changed("port") {
				if useHTTPS {
					port = 443
				} else {
					port = 80
				}
			}
			uploadURL = fmt.Sprintf("%s://%s:%d/upload/%s", scheme, host, port, token)
		}
		excludeDirs, _ := cmd.Flags().GetStringSlice("exclude-dir")
		return client.SendFiles(uploadURL, args, excludeDirs)
	},
}

func init() {
	sendCmd.Flags().StringP("host", "H", "", "Server host or IP address")
	sendCmd.Flags().IntP("port", "p", 0, "Server port (default 80 for http, 443 for https)")
	sendCmd.Flags().Bool("http", false, "Use HTTP (default)")
	sendCmd.Flags().Bool("https", false, "Use HTTPS")
	sendCmd.Flags().StringP("token", "t", "", "Auth token (required)")
	sendCmd.Flags().String("relay", "", "Relay server URL (send via cloud relay instead of direct connection)")
	sendCmd.Flags().StringSlice("exclude-dir", nil, "Directory names to exclude (can be repeated or comma-separated)")
	sendCmd.MarkFlagsMutuallyExclusive("http", "https")
}
