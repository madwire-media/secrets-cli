package vault

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/madwire-media/secrets-cli/util"
	"github.com/madwire-media/secrets-cli/vars"
)

var auth vaultController

type vaultController struct {
	init            bool
	config          vaultConfig
	validatedTokens map[string]struct{}
}

type vaultConfig struct {
	TokenCache map[string]cachedToken `json:"tokenCache"`
}

type cachedToken struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

func (controller *vaultController) Init() error {
	if controller.init {
		return nil
	}

	config := vaultConfig{
		TokenCache: make(map[string]cachedToken),
	}

	if !vars.IsCICD {
		err := util.LoadConfig("vault", &config)
		if err != nil {
			return err
		}

		if config.TokenCache == nil {
			config.TokenCache = make(map[string]cachedToken)
		}

		needToSave := false
		now := time.Now().Unix()

		for key, cached := range config.TokenCache {
			if cached.Expires < now {
				delete(config.TokenCache, key)
				needToSave = true
			}
		}

		if needToSave {
			err = controller.save()
			if err != nil {
				return err
			}
		}
	}

	controller.init = true
	controller.config = config
	controller.validatedTokens = make(map[string]struct{})

	return nil
}

func (controller *vaultController) PrepareForURL(parsedURL *url.URL) error {
	err := controller.Init()
	if err != nil {
		return err
	}

	optToken, vaultAuth, err := ensureAuthConfiguredForURL(parsedURL)
	if err != nil {
		return err
	}

	if optToken != nil {
		key, shouldCache, err := cacheKeyForAuth(parsedURL.Host, vaultAuth)
		if err != nil {
			return err
		}

		controller.validatedTokens[key] = struct{}{}

		if shouldCache {
			cached := cachedToken{
				Token:   *optToken,
				Expires: time.Now().Unix() + 30*24*60*60,
			}
			controller.config.TokenCache[key] = cached

			err := controller.save()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (controller *vaultController) GetTokenForURL(parsedURL *url.URL) (string, error) {
	vaultAuth, ok := (*vars.Auth.Vault)[parsedURL.Host]

	if !ok {
		return "", errors.New("Host not configured, was PrepareForUrl called?")
	}

	key, shouldCache, err := cacheKeyForAuth(parsedURL.Host, &vaultAuth)

	if cached, ok := controller.config.TokenCache[key]; ok {
		if _, ok = controller.validatedTokens[key]; ok {
			return cached.Token, nil
		}

		valid := validateTokenForURL(parsedURL, cached.Token)
		if valid {
			controller.validatedTokens[key] = struct{}{}

			cached.Expires = time.Now().UTC().Unix() + 30*24*60*60
			err := controller.save()
			if err != nil {
				return "", err
			}

			return cached.Token, nil
		}

		delete(controller.config.TokenCache, key)
	}

	token, err := getTokenForURL(parsedURL, &vaultAuth)
	if err != nil {
		return "", err
	}

	controller.validatedTokens[key] = struct{}{}

	if shouldCache {
		cached := cachedToken{
			Token:   token,
			Expires: time.Now().UTC().Unix() + 30*24*60*60,
		}

		controller.config.TokenCache[key] = cached

		err = controller.save()
		if err != nil {
			return "", err
		}
	}

	return token, nil
}

func (controller *vaultController) save() error {
	if vars.IsCICD {
		return nil
	}

	return util.SaveConfig("vault", &controller.config)
}

func validateTokenForURL(parsedURL *url.URL, token string) bool {
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	_, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return false
	}

	return resp.StatusCode == 200
}
