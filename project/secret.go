package project

import (
	"errors"

	"github.com/madwire-media/secrets-cli/engines/vault"
	"github.com/madwire-media/secrets-cli/types"
)

// SecretConfig is the format for any secret in the secrets.yaml config
type SecretConfig struct {
	File  string              `yaml:"file"`
	Class *string             `yaml:"class,omitempty"`
	Vault *vault.SecretConfig `yaml:"vault,omitempty"`
}

// Prepare prepares this secret for fetching, for example by getting auth
// credentials from the user and generating a valid login session
func (secretConfig *SecretConfig) Prepare() error {
	if secretConfig.Vault != nil {
		return secretConfig.Vault.Prepare()
	}

	return errors.New("No secret engine defined for secret")
}

// Fetch fetches this secret from the remote server, returning an interface
// capable of uploading a new version of this secret as well
func (secretConfig *SecretConfig) Fetch() (types.FetchedSecret, error) {
	if secretConfig.Vault != nil {
		return secretConfig.Vault.Fetch()
	}

	return nil, errors.New("No secret engine defined for secret")
}
