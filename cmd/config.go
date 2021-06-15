package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Collection of config subcommands",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
