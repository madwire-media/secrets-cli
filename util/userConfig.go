package util

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"runtime"
)

// GetConfigDir gets the os-dependent user configuration directory
func GetConfigDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "linux", "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir = home + "/.config/mw-secrets/"

	case "windows":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir = home + "\\AppData\\Roaming\\mw-secrets\\"
	}

	info, err := os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0770)
		if err != nil {
			return "", err
		}
	} else if !info.IsDir() {
		return "", errors.New("Config path is not a directory")
	}

	return dir, nil
}

// LoadConfig loads a user config file of the given name
func LoadConfig(config string, data interface{}) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	filename := dir + config + ".json"

	text, err := ioutil.ReadFile(filename)
	if err != nil {
		text = []byte("{}")
	}

	err = json.Unmarshal(text, data)
	if err != nil {
		return err
	}

	return nil
}

// SaveConfig saves a user config file of the given name
func SaveConfig(config string, data interface{}) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	filename := dir + config + ".json"

	text, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, text, 0660)
	if err != nil {
		return err
	}

	return nil
}

// LoadExternalConfig reads an external config file, outside of the default user
// config directory
func LoadExternalConfig(filename string, data interface{}) error {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(text, data)
	if err != nil {
		return err
	}

	return nil
}
