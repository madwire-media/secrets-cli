package cmd

import (
	"fmt"
	"os"

	"github.com/madwire-media/secrets-cli/types"
	"github.com/madwire-media/secrets-cli/util"
	"github.com/madwire-media/secrets-cli/vars"
	"github.com/spf13/cobra"
)

var authFiles []string

var rootCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Tracks local secret files to Vault secrets",
}

func init() {
	cobra.OnInitialize(initSettings, autoUpdate)
	rootCmd.PersistentFlags().StringArrayVar(&authFiles, "auth-config", []string{}, "one or more auth config files")
	rootCmd.PersistentFlags().BoolVar(&vars.IsCICD, "cicd", false, "shortcut to streamline settings for CI/CD usage")
}

func initSettings() {
	// Don't do TTY things when CI/CD flag is enabled
	if vars.IsCICD {
		vars.IsTTY = false
	}

	// Load user auth if CI/CD flag is not enabled
	if !vars.IsCICD {
		err := util.LoadUserAuth()
		if err != nil {
			panic(err)
		}

		util.MergeAuth(&vars.Auth, &vars.UserAuth)
	}

	// Load extra auth configs
	for _, file := range authFiles {
		var auth types.RootAuth

		err := util.LoadExternalConfig(file, &auth)
		if err != nil {
			panic(err)
		}

		util.MergeAuth(&vars.Auth, &auth)
	}

	// Set flag if we have user auth only
	if !vars.IsCICD && len(authFiles) == 0 {
		vars.UserAuthOnly = true
	}
}

func autoUpdate() {
	// Try updating if we aren't in CI/CD mode
	if !vars.IsCICD {
		if err := util.TryAutoUpdateSelf(); err != nil {
			fmt.Println(err.Error())
		}
	}
}

// Execute executes the root command. This should only be called by the
// program's main function
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
