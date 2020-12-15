package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/madwire-media/secrets-cli/types"
	"github.com/madwire-media/secrets-cli/util"
)

const (
	casMismatch = "check-and-set parameter did not match the current version"
)

// FetchedVaultSecret is an implementation of types.FetchedSecret specifically
// for a secret fetched from Vault
type FetchedVaultSecret struct {
	value         interface{}
	version       int
	format        int
	isMissingData bool

	apiURL  *url.URL
	mapping Mapping
}

// Value returns the fetched secret value or sub-value
func (fetched *FetchedVaultSecret) Value() interface{} {
	return fetched.value
}

// Version returns the fetched secret's version
func (fetched *FetchedVaultSecret) Version() interface{} {
	return fetched.version
}

// Format returns the configured format for this secret
func (fetched *FetchedVaultSecret) Format() int {
	return fetched.format
}

// IsMissingData returns true if the remote secret existed but was incomplete
func (fetched *FetchedVaultSecret) IsMissingData() bool {
	return fetched.isMissingData
}

// UploadNew modifies the remote secret and replaces the value or sub-value with
// a new given value, and returns the new secret version
func (fetched *FetchedVaultSecret) UploadNew(value interface{}) (interface{}, error) {
	type secretPost struct {
		Options struct {
			CAS *int `json:"cas"`
		} `json:"options"`
		Data map[string]interface{} `json:"data"`
	}

	type okResponse struct {
		Data struct {
			Version int `json:"version"`
		} `json:"data"`
	}

	type errorResponse struct {
		Data struct {
			Error string `json:"error"`
		} `json:"data"`
	}

	token, err := auth.GetTokenForURL(fetched.apiURL)
	if err != nil {
		return nil, err
	}

	for {
		// Get the latest secret data and version
		req, _ := http.NewRequest("GET", fetched.apiURL.String(), nil)
		req.Header.Add("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var cas int
		var data map[string]interface{}

		if resp.StatusCode == 404 {
			cas = 0
			data = make(map[string]interface{})
		} else if resp.StatusCode == 200 {
			rawSecretData := rawSecret{}
			err = json.Unmarshal(body, &rawSecretData)
			if err != nil {
				return nil, err
			}

			cas = rawSecretData.Data.Metadata.Version
			data = rawSecretData.Data.Data
		} else {
			return nil, errors.New("Got status " + resp.Status + " while fetching secret for push")
		}

		var dataAsInterface interface{} = data

		// Modify the secret based on the mapping
		if fetched.mapping.FromData != nil {
			err = util.SetAtPath(&dataAsInterface, fetched.mapping.FromData.Path, value)
		} else if fetched.mapping.FromText != nil {
			err = util.SetAtPath(&dataAsInterface, &fetched.mapping.FromText.Path, value)
		}

		if err != nil {
			return nil, err
		}

		// Upload the modified secret with check-and-set
		var postData secretPost

		postData.Options.CAS = &cas
		postData.Data = dataAsInterface.(map[string]interface{})

		postBytes, err := json.Marshal(postData)
		if err != nil {
			return nil, err
		}

		req, _ = http.NewRequest("POST", fetched.apiURL.String(), bytes.NewReader(postBytes))
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)

		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 200 {
			// check-and-set succeeded, data was written, return the new version

			okData := okResponse{}
			err = json.Unmarshal(body, &okData)
			if err != nil {
				return nil, err
			}

			return okData.Data.Version, nil
		}

		errorData := errorResponse{}
		err = json.Unmarshal(body, &errorData)
		if err != nil {
			return nil, errors.New("Got status " + resp.Status + " while setting secret")
		}

		if errorData.Data.Error != casMismatch {
			return nil, errors.New("Got status " + resp.Status + " while setting secret")
		}

		// If there was a CAS mismatch, then do the whole thing over again
		fmt.Println("info: remote secret was edited during push, retrying")
	}
}

// SecretConfig contains the Vault-specific configuration parameters for a
// secret in secrets.yaml
type SecretConfig struct {
	URL     string  `yaml:"url"`
	Mapping Mapping `yaml:"mapping"`
}

// Mapping represents a data or text mapping of a Vault key/value secret
// document
type Mapping struct {
	FromData *struct {
		Format string         `yaml:"format"`
		Path   *[]interface{} `yaml:"path"`
	} `yaml:"fromData"`
	FromText *struct {
		Path []interface{} `yaml:"path"`
	} `yaml:"fromText"`
}

type rawSecret struct {
	Data struct {
		Data     map[string]interface{} `json:"data"`
		Metadata struct {
			CreatedTime  string `json:"created_time"`
			DeletionTime string `json:"deletion_time"`
			Destroyed    bool   `json:"destroyed"`
			Version      int    `json:"version"`
		} `json:"metadata"`
	} `json:"data"`
}

// Prepare ensures that the Vault engine has all the required authentication
// parameters to fetch this secret
func (secretConfig *SecretConfig) Prepare() error {
	parsedURL, err := url.Parse(secretConfig.URL)
	if err != nil {
		return err
	}

	return auth.PrepareForURL(parsedURL)
}

// Fetch downloads this secret and returns an instance of FetchedVaultSecret
func (secretConfig *SecretConfig) Fetch() (types.FetchedSecret, error) {
	parsedURL, err := url.Parse(secretConfig.URL)
	if err != nil {
		return nil, err
	}

	var secret FetchedVaultSecret

	// Compute the format first in case of an early exit (i.e. 404)
	if secretConfig.Mapping.FromData != nil {
		switch secretConfig.Mapping.FromData.Format {
		case "json":
			secret.format = util.FormatJSON
		case "yaml":
			secret.format = util.FormatYaml
		default:
			return nil, errors.New("Unknown format")
		}
	} else if secretConfig.Mapping.FromText != nil {
		secret.format = util.FormatText

		if len(secretConfig.Mapping.FromText.Path) == 0 {
			return nil, errors.New("No path provided for fromText secret mapping")
		}
	} else {
		return nil, errors.New("No mapping provided for secret")
	}

	secret.mapping = secretConfig.Mapping

	// Insert /data into the secret path after the secrets engine, assuming k/v
	// engine v2
	path := parsedURL.Path
	idx := strings.IndexRune(path[1:], '/')
	if idx == -1 {
		return nil, errors.New("URL has only one path segment")
	}
	idx++

	parsedURL.Path = "/v1" + path[:idx] + "/data" + path[idx:]
	secret.apiURL = parsedURL

	// Get the secret data
	token, err := auth.GetTokenForURL(parsedURL)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", parsedURL.String(), nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		secret.isMissingData = true
		return &secret, nil
	} else if resp.StatusCode != 200 {
		return nil, errors.New("Got status " + resp.Status + " while fetching secret")
	}

	rawSecretData := rawSecret{}
	err = json.Unmarshal(body, &rawSecretData)
	if err != nil {
		return nil, err
	}

	secret.version = rawSecretData.Data.Metadata.Version

	// Process secret data
	if secretConfig.Mapping.FromData != nil {
		var data interface{}

		if secretConfig.Mapping.FromData.Path != nil {
			data, err = util.TraversePath(rawSecretData.Data.Data, secretConfig.Mapping.FromData.Path)
			if err != nil {
				if util.IsMissingData(err) {
					secret.isMissingData = true
					return &secret, nil
				}

				return nil, err
			}
		} else {
			data = rawSecretData.Data.Data
		}

		secret.value = data
	} else if secretConfig.Mapping.FromText != nil {
		data, err := util.TraversePath(rawSecretData.Data.Data, &secretConfig.Mapping.FromText.Path)
		if err != nil {
			if util.IsMissingData(err) {
				secret.isMissingData = true
				return &secret, nil
			}

			return nil, err
		}

		switch v := data.(type) {
		case string:
			secret.value = v
		default:
			return nil, errors.New("Value for text mapping is not a string")
		}
	}

	return &secret, nil
}

// TODO: for future, use this code for example in grabbing Vault secret
// https://github.com/hashicorp/consul-template/blob/7eebde9030a600fec83820fad6cfed2f9ecbd77c/dependency/vault_read.go#L36
