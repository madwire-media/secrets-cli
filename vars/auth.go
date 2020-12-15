package vars

import "github.com/madwire-media/secrets-cli/types"

var (
	// Auth is the global computed authentication config
	Auth types.RootAuth = makeRootAuth()

	// UserAuth is the auth config stored only in the user config dir
	UserAuth types.RootAuth = makeRootAuth()

	// UserAuthOnly is true when no outside auth files are loaded or provided on
	// the CLI args
	UserAuthOnly bool = true

	// UserAuthLoaded is true when the default user auth has been loaded
	UserAuthLoaded bool = false
)

func makeRootAuth() types.RootAuth {
	vault := make(map[string]types.VaultAuth)

	return types.RootAuth{
		Vault: &vault,
	}
}
