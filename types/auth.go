package types

// RootAuth holds auth configurations for every secrets engine
type RootAuth struct {
	Vault *map[string]VaultAuth `json:"vault"`
}
