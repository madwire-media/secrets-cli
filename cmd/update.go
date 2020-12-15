package cmd

import (
	"fmt"

	"github.com/madwire-media/secrets-cli/util"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Performs an update check",
	Long:  "Checks GitHub for a new release and updates itself if there is one",
	Run: func(cmd *cobra.Command, args []string) {
		err := util.TryManualUpdate()
		if err != nil {
			fmt.Println("Error during update:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
