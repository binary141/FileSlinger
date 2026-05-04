package cmd

import (
	"fileSlinger/client"

	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send [files...]",
	Short: "Send one or more files to a fileslinger server",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		return client.SendFiles(host, port, args)
	},
}

func init() {
	sendCmd.Flags().StringP("host", "H", "", "Server host or IP address (required)")
	sendCmd.Flags().IntP("port", "p", 8080, "Server port")
	sendCmd.MarkFlagRequired("host")
}
