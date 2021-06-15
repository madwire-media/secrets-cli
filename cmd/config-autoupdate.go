package cmd

import (
	"fmt"
	"os"

	"github.com/madwire-media/secrets-cli/util"
	"github.com/spf13/cobra"
)

var configAutoupdateCmd = &cobra.Command{
	Use:   "autoupdate",
	Short: "Configure automatic updates",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		yes, _ := cmd.Flags().GetBool("yes")
		no, _ := cmd.Flags().GetBool("no")

		var enabled bool

		if yes {
			enabled = true
		} else if no {
			enabled = false
		} else {
			alreadyEnabled, err := util.GetAutoUpdate()
			if err != nil {
				fmt.Println("Error reading user config:", err)
				os.Exit(1)
				return
			}

			var stateString string

			if alreadyEnabled {
				stateString = "enabled"
			} else {
				stateString = "disabled"
			}

			enabled = util.CliQuestionYesNo("Enable automatic updates? (currently " + stateString + ")")
		}

		changed, err := util.SetAutoUpdate(enabled)
		if err != nil {
			fmt.Println("Error saving user config:", err)
			os.Exit(1)
			return
		}

		if changed {
			fmt.Println("Config saved")
		} else {
			fmt.Println("Config unchanged")
		}
	},
}

func init() {
	configCmd.AddCommand(configAutoupdateCmd)

	configAutoupdateCmd.Flags().Bool("yes", false, "Enable automatic updates")
	configAutoupdateCmd.Flags().Bool("no", false, "Disable automatic updates")
}
