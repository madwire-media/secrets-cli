package cmd

import (
	"fmt"

	"github.com/madwire-media/secrets-cli/vars"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the current build version and commit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Build Version:", vars.BuildVersion)
		fmt.Println("Build Commit:", vars.BuildCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
