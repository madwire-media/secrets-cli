package util

import (
	"github.com/madwire-media/secrets-cli/types"
	"github.com/madwire-media/secrets-cli/vars"
)

// MergeAuth merges an overlay RootAuth into a base RootAuth
func MergeAuth(base *types.RootAuth, overlay *types.RootAuth) {
	if overlay.Vault != nil {
		if base.Vault == nil {
			base.Vault = overlay.Vault
		} else {
			for key, value := range *overlay.Vault {
				(*base.Vault)[key] = value
			}
		}
	}
}

// LoadUserAuth loads the default user auth config
func LoadUserAuth() error {
	var auth types.RootAuth

	err := LoadConfig("auth", &auth)
	if err != nil {
		return err
	}

	MergeAuth(&vars.UserAuth, &auth)

	vars.UserAuthLoaded = true

	return nil
}

// SaveUserAuth saves the default user auth config
func SaveUserAuth() error {
	return SaveConfig("auth", &vars.UserAuth)
}
