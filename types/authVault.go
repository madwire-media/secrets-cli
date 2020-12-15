package types

// VaultAuth holds any valid, implemented Vault authentication method
type VaultAuth struct {
	Userpass *VaultAuthUserpass `json:"userpass,omitempty"`
	AppRole  *VaultAuthAppRole  `json:"appRole,omitempty"`
	Token    *string            `json:"token,omitempty"`
}

// VaultAuthUserpass holds a username and password for userpass authentication
type VaultAuthUserpass struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// VaultAuthAppRole holds a role ID and secret ID for authrole authentication
type VaultAuthAppRole struct {
	RoleID   string `json:"roleID"`
	SecretID string `json:"secretID"`
}
