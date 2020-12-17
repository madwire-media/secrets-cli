package project

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/madwire-media/secrets-cli/types"
	"github.com/madwire-media/secrets-cli/util"
	"github.com/madwire-media/secrets-cli/vars"
	"github.com/ryanuber/go-glob"
	"gopkg.in/yaml.v3"
)

// LockState represents the entire state of a lockfile and every current secret
type LockState struct {
	Files map[string]LockedFile `yaml:"files"`
}

// LockedFile represents the state of a particular secret file
type LockedFile struct {
	RemoteVersion interface{} `yaml:"remoteVersion,omitempty"`
	LocalHash     string      `yaml:"localHash,omitempty"`
	LocalFormat   string      `yaml:"localFormat"`
	formatError   error
	data          interface{}
}

func (project *Project) readLockfile() error {
	filename := filepath.Join(project.path, "secrets.lock")

	lockBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}

	state := LockState{
		Files: make(map[string]LockedFile),
	}

	err = yaml.Unmarshal(lockBytes, &state)
	if err != nil {
		return err
	}

	project.lastState = state

	return nil
}

func (project *Project) saveCurrentState() error {
	err := project.ensureLockfileInGitignore()
	if err != nil {
		return err
	}

	filename := filepath.Join(project.path, "secrets.lock")

	lockBytes, err := yaml.Marshal(&project.currentState)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, lockBytes, 0777)
	if err != nil {
		return err
	}

	return nil
}

func (project *Project) computeCurrentState(
	secrets []SecretConfig,
	fetched []types.FetchedSecret,
) error {
	project.currentState.Files = make(map[string]LockedFile)

	for i := 0; i < len(secrets); i++ {
		secret := secrets[i]
		fetchedSecret := fetched[i]

		correctedFilename := filepath.Join(project.path, secret.File)
		prevState, hasPrevState := project.lastState.Files[secret.File]

		fileState := LockedFile{
			RemoteVersion: fetchedSecret.Version(),
		}

		bytes, err := ioutil.ReadFile(correctedFilename)
		if err == nil {
			var parsed interface{}
			var err error
			var localFormat string
			parsedSuccessfully := false

			// Try the format in the lock file first
			if hasPrevState {
				format := util.NameToFormat(prevState.LocalFormat)

				data, err := util.ParseData(bytes, format)
				if err != nil {
					return err
				}

				if err == nil {
					parsedSuccessfully = true
					parsed = data
					localFormat = prevState.LocalFormat
				}
			}

			// Try the new format specified by the user config as a fallback
			if !parsedSuccessfully {
				parsed, err = util.ParseData(bytes, fetchedSecret.Format())
				localFormat = util.FormatToName(fetchedSecret.Format())
			}

			if err != nil {
				fileState.formatError = err
			} else {
				hash, err := hashValue(&parsed)
				if err != nil {
					return err
				}

				fileState.LocalHash = hash
				fileState.LocalFormat = localFormat
				fileState.data = parsed
			}
		}

		project.currentState.Files[secret.File] = fileState
	}

	return nil
}

func (project *Project) ensureLockfileInGitignore() error {
	// Don't worry about changing the .gitignore in CI/CD mode
	if vars.IsCICD {
		return nil
	}

	filename := filepath.Join(project.path, ".gitignore")

	gitignore, err := ioutil.ReadFile(filename)
	var newGitignore string

	if os.IsNotExist(err) {
		newGitignore = "/secrets.lock\n"
	} else {
		scanner := bufio.NewScanner(bytes.NewReader(gitignore))
		for scanner.Scan() {
			text := scanner.Text()

			hashIdx := strings.Index(text, "#")
			if hashIdx > 0 {
				text = text[hashIdx:]
			}

			text = strings.TrimSpace(text)
			text = strings.TrimLeft(text, "/")

			// .gitignore already has matching line for secrets.lock
			if glob.Glob(text, "secrets.lock") {
				return nil
			}
		}

		newGitignore = string(gitignore)

		if newGitignore[len(newGitignore)-1:] != "\n" {
			newGitignore += "\n"
		}

		newGitignore += "/secrets.lock\n"
	}

	return ioutil.WriteFile(filename, []byte(newGitignore), 0664)
}

func hashValue(value interface{}) (string, error) {
	asJSON, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	digest := sha256.Sum256(asJSON)
	return hex.EncodeToString(digest[:]), nil
}
