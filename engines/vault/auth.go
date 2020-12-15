package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/madwire-media/secrets-cli/types"
	"github.com/madwire-media/secrets-cli/util"
	"github.com/madwire-media/secrets-cli/vars"
)

func cacheKeyForAuth(host string, vaultAuth *types.VaultAuth) (string, bool, error) {
	if vaultAuth.Token != nil {
		key := fmt.Sprintf("%s,token,%s", host, *vaultAuth.Token)
		return key, false, nil
	}

	if vaultAuth.AppRole != nil {
		key := fmt.Sprintf("%s,approle,%s", host, vaultAuth.AppRole.RoleID)
		return key, true, nil
	}

	if vaultAuth.Userpass != nil {
		key := fmt.Sprintf("%s,userpass,%s", host, vaultAuth.Userpass.Username)
		return key, true, nil
	}

	return "", false, errors.New("Auth config empty for host " + host)
}

func ensureAuthConfiguredForURL(parsedURL *url.URL) (*string, *types.VaultAuth, error) {
	var configForHost *types.VaultAuth

	if vars.Auth.Vault != nil {
		if config, ok := (*vars.Auth.Vault)[parsedURL.Host]; ok {
			configForHost = &config
		}
	}

	if configForHost == nil || !hasAnyAuth(configForHost) {
		var shouldCreateConfig bool

		if !vars.IsTTY {
			shouldCreateConfig = false
		} else if vars.UserAuthOnly {
			fmt.Printf("No auth config for Vault instance at '%s', please enter a username and password to be saved locally\n", parsedURL.Host)
			shouldCreateConfig = true
		} else if vars.UserAuthLoaded {
			fmt.Printf("No auth config for Vault instance at '%s', would you like to enter a username and password to be saved locally in your home directory?\n", parsedURL.Host)
			shouldCreateConfig = util.CliQuestionYesNoDefault("Create local login config?", false)
		} else {
			shouldCreateConfig = false
		}

		if !shouldCreateConfig {
			return nil, nil, fmt.Errorf("No auth config for Vault instance at '%s'", parsedURL.Host)
		}

		for {
			username := util.CliQuestion("Username")
			password, err := util.CliQuestionHidden("Password")
			if err != nil {
				return nil, nil, err
			}

			vaultAuth := types.VaultAuth{
				Userpass: &types.VaultAuthUserpass{
					Username: username,
					Password: password,
				},
			}

			token, err := getTokenForURL(parsedURL, &vaultAuth)
			if err != nil {
				fmt.Printf("Error, please try again: %s\n", err.Error())
			} else {
				(*vars.UserAuth.Vault)[parsedURL.Host] = vaultAuth
				(*vars.Auth.Vault)[parsedURL.Host] = vaultAuth

				err = util.SaveUserAuth()
				if err != nil {
					return nil, nil, err
				}

				return &token, &vaultAuth, nil
			}
		}
	}

	return nil, configForHost, nil
}

func hasAnyAuth(vaultAuth *types.VaultAuth) bool {
	if vaultAuth.Token != nil {
		return true
	}

	if vaultAuth.AppRole != nil {
		return true
	}

	if vaultAuth.Userpass != nil {
		return true
	}

	return false
}

func getTokenForURL(parsedURL *url.URL, vaultAuth *types.VaultAuth) (string, error) {
	if vaultAuth.Token != nil {
		return *vaultAuth.Token, nil
	}

	if vaultAuth.AppRole != nil {
		return getTokenForURLWithAppRole(parsedURL, vaultAuth.AppRole)
	}

	if vaultAuth.Userpass != nil {
		return getTokenForURLWithUserpass(parsedURL, vaultAuth.Userpass)
	}

	return "", errors.New("Auth config exists but is empty for " + parsedURL.Host)
}

func getTokenForURLWithUserpass(parsedURL *url.URL, userpass *types.VaultAuthUserpass) (string, error) {
	type postData struct {
		Password string `json:"password"`
	}

	type loginResponse struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}

	loginURL := *parsedURL
	loginURL.Path = "/v1/auth/userpass/login/" + userpass.Username
	loginURL.Fragment = ""

	postBytes, err := json.Marshal(postData{
		Password: userpass.Password,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(loginURL.String(), "application/json", bytes.NewReader(postBytes))
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	login := loginResponse{}
	err = json.Unmarshal(body, &login)
	if err != nil {
		return "", err
	}

	if login.Auth.ClientToken == "" {
		return "", errors.New("Login failed: " + resp.Status)
	}

	return login.Auth.ClientToken, nil
}

func getTokenForURLWithAppRole(parsedURL *url.URL, appRole *types.VaultAuthAppRole) (string, error) {
	type postData struct {
		RoleID   string `json:"role_id"`
		SecretID string `json:"secret_id"`
	}

	type loginResponse struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}

	loginURL := *parsedURL
	loginURL.Path = "/v1/auth/approle/login"
	loginURL.Fragment = ""

	postBytes, err := json.Marshal(postData{
		RoleID:   appRole.RoleID,
		SecretID: appRole.SecretID,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(loginURL.String(), "application/json", bytes.NewReader(postBytes))
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	login := loginResponse{}
	err = json.Unmarshal(body, &login)
	if err != nil {
		return "", err
	}

	if login.Auth.ClientToken == "" {
		return "", errors.New("Login failed: " + resp.Status)
	}

	return login.Auth.ClientToken, nil
}
