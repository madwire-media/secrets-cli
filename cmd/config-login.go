package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/madwire-media/secrets-cli/engines/vault"
	"github.com/madwire-media/secrets-cli/types"
	"github.com/madwire-media/secrets-cli/util"
	"github.com/madwire-media/secrets-cli/vars"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var configLoginCmd = &cobra.Command{
	Use:   "login [host]",
	Short: "Login to a Vault server",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var host string

		if len(args) == 1 {
			host = args[0]
		} else {
			if !vars.IsTTY {
				fmt.Println("Error: must specify host as an argument or use a TTY")
				os.Exit(1)
				return
			}

			host = util.CliQuestion("Vault host (domain and port only)")
		}

		vaultAuth, err := getLoginAuth(host, cmd.Flags())
		if err != nil {
			fmt.Println("Error getting login auth:", err)
			os.Exit(1)
			return
		}

		err = vault.TryAuth(host, vaultAuth)
		if err != nil {
			fmt.Println("Error logging in to Vault:", err)
			os.Exit(1)
			return
		}

		saveTo, _ := cmd.Flags().GetString("save-to")

		if saveTo == "" {
			err = util.LoadUserAuth()
			if err != nil {
				fmt.Println("Error loading user auth config:", err)
				os.Exit(1)
				return
			}

			util.MergeAuth(&vars.UserAuth, &types.RootAuth{
				Vault: &map[string]types.VaultAuth{
					host: *vaultAuth,
				},
			})

			err = util.SaveUserAuth()
			if err != nil {
				fmt.Println("Error saving user auth config:", err)
				os.Exit(1)
				return
			}
		} else {
			var auth types.RootAuth

			// If this fails, just assume the file doesn't exist and we're
			// creating a new one
			util.LoadExternalConfig(saveTo, &auth)

			util.MergeAuth(&auth, &types.RootAuth{
				Vault: &map[string]types.VaultAuth{
					host: *vaultAuth,
				},
			})

			err = util.SaveExternalConfig(saveTo, &auth)
			if err != nil {
				fmt.Println("Error saving external auth config:", err)
				os.Exit(1)
				return
			}
		}

		fmt.Println("Login succeeded, config saved")
	},
}

func init() {
	configCmd.AddCommand(configLoginCmd)

	configLoginCmd.Flags().String("save-to", "", "Where to save the login info (defaults to user auth file)")
	configLoginCmd.Flags().String("token", "", "Vault token")
	configLoginCmd.Flags().String("username", "", "username for userpass auth")
	configLoginCmd.Flags().String("password", "", "password for userpass auth (optional)")
	configLoginCmd.Flags().String("role-id", "", "role ID for AppRole auth")
	configLoginCmd.Flags().String("secret-id", "", "secret ID for AppRole auth (optional)")
	configLoginCmd.Flags().Bool("oidc", false, "Use OIDC auth method")
	configLoginCmd.Flags().String("oidc-mount", "", "OIDC mount path")
}

func getLoginAuth(host string, flags *pflag.FlagSet) (*types.VaultAuth, error) {
	token, _ := flags.GetString("token")
	username, _ := flags.GetString("username")
	password, _ := flags.GetString("password")
	roleID, _ := flags.GetString("role-id")
	secretID, _ := flags.GetString("secret-id")
	oidc, _ := flags.GetBool("oidc")
	oidcMount, _ := flags.GetString("oidc-mount")

	var auth types.VaultAuth
	var err error

	if token != "" {
		auth.Token = &token
	} else if username != "" {
		if password == "" {
			if !vars.IsTTY {
				return nil, errors.New("must specify --password or use a TTY")
			}

			password, err = util.CliQuestionHidden("Password")
			if err != nil {
				return nil, err
			}
		}

		auth.Userpass = &types.VaultAuthUserpass{
			Username: username,
			Password: password,
		}
	} else if password != "" {
		return nil, errors.New("password is defined but no username is defined")
	} else if roleID != "" {
		if secretID != "" {
			if !vars.IsTTY {
				return nil, errors.New("must specify --secret-id or use a TTY")
			}

			secretID = util.CliQuestion("Secret ID")
		}

		auth.AppRole = &types.VaultAuthAppRole{
			RoleID:   roleID,
			SecretID: secretID,
		}
	} else if secretID != "" {
		return nil, errors.New("role-id is defined but no secret-id is defined")
	} else if oidc {
		if vars.IsCICD {
			return nil, errors.New("OIDC auth not supported in CI/CD mode")
		}

		if oidcMount == "" && vars.IsTTY {
			oidcMount = util.CliQuestion("OIDC mount path (defaults to \"oidc\")")
		}

		auth.OIDC = &types.VaultAuthOIDC{
			Mount: oidcMount,
		}
	} else {
		if !vars.IsTTY {
			return nil, errors.New("must specify credentials as arguments or use a TTY")
		}

		return vault.GetLoginAuthTTY()
	}

	return &auth, nil
}
