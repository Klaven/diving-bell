package cmd

import (
	"github.com/spf13/cobra"

	divingbell "github.com/tdaines42/diving-bell/pkg/diving-bell"
)

func init() {
	configCmd.AddCommand(configRetrieveCmd)
}

var configRetrieveCmd = &cobra.Command{
	Use:   "retrieve [cluster name]",
	Short: "Retrieve the config from the cluster and print it to console",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		divingbell.RetrieveClusterConfig(args[0])
	},
}
