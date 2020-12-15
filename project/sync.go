package project

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/madwire-media/secrets-cli/types"
	"github.com/madwire-media/secrets-cli/util"
	"github.com/madwire-media/secrets-cli/vars"
)

// SyncOptions contains every relevant option for the sync operation, mostly
// intended to be set with CLI flags
type SyncOptions struct {
	PullByDefault bool
	PushByDefault bool
	FixByDefault  bool
	Classes       ClassUpdate
}

// Sync prepares every secret, fetches every secret, and does a 3-way diff
// between the local files, the local lock file, and the remote secrets,
// pulling or pushing secrets where relevant
func (project *Project) Sync(options SyncOptions) error {
	if options.PullByDefault && options.PushByDefault {
		return errors.New("--pull flag and --push flag cannot both be enabled")
	}

	project.applyClassUpdate(options.Classes)

	secrets, excludedSecrets := filterSecrets(project.Secrets, project.classes)

	for _, secret := range secrets {
		err := secret.Prepare()
		if err != nil {
			return err
		}
	}

	fetchedSecrets := make([]types.FetchedSecret, len(secrets))

	for idx, secret := range secrets {
		fetchedSecret, err := secret.Fetch()
		if err != nil {
			return err
		}

		fetchedSecrets[idx] = fetchedSecret
	}

	err := project.computeCurrentState(secrets, fetchedSecrets)
	if err != nil {
		return err
	}

	for idx, secret := range secrets {
		fetchedSecret := fetchedSecrets[idx]
		fileState := project.currentState.Files[secret.File]
		prevState, hasPrevState := project.lastState.Files[secret.File]

		correctedFilename := filepath.Join(project.path, secret.File)
		relativeFilename, err := filepath.Rel(vars.Workdir, correctedFilename)
		if err != nil {
			return err
		}

		remoteHash, err := hashValue(fetchedSecret.Value())
		if err != nil {
			return err
		}

		if fileState.LocalHash == "" && fileState.formatError == nil {
			// The file doesn't exist

			if fetchedSecret.IsMissingData() {
				// Neither remote secret nor local file exist

				fmt.Printf("No local file or remote data for secret '%s'\n", relativeFilename)
			} else {
				// Remote secret exists, but local file doesn't

				fmt.Printf("Writing new secret to '%s'\n", relativeFilename)

				err := project.pullSecret(secret, fetchedSecret, &fileState)
				if err != nil {
					return err
				}

				fmt.Println("    done")
			}
		} else if fetchedSecret.IsMissingData() {
			// File exists but remote secret does not

			shouldPush := false

			if vars.IsCICD {
				return fmt.Errorf("Remote secret for '%s' does not exist or is missing data", relativeFilename)
			} else if options.FixByDefault {
				fmt.Printf("Pushing secret '%s' because remote secret is incomplete or does not exist (--fix flag is enabled)\n", relativeFilename)
				shouldPush = true
			} else if !vars.IsTTY {
				fmt.Printf("Remote secret for '%s' is incomplete or does not exist, use the --fix flag to fix it\n", relativeFilename)
			} else {
				fmt.Printf("Remote secret for '%s' is incomplete or does not exist, do you want to push it?\n", relativeFilename)
				shouldPush = util.CliQuestionYesNoDefault("Push secret?", true)
			}

			if shouldPush {
				err := project.pushSecret(fetchedSecret, &fileState)
				if err != nil {
					return err
				}

				fmt.Println("    done")
			} else {
				fmt.Println("    skipped")
			}
		} else if fileState.LocalHash == remoteHash {
			// File exists and contents match existing secret

			if util.NameToFormat(fileState.LocalFormat) != fetchedSecret.Format() {
				// Local file format does not match expected format

				shouldPull := false

				if hasPrevState && prevState.LocalFormat == fileState.LocalFormat {
					fmt.Printf("Updating secret '%s' to newer format\n", relativeFilename)
					shouldPull = true
				} else if vars.IsCICD {
					fmt.Printf("Updating secret '%s' to correct format (--cicd flag is enabled)\n", relativeFilename)
					shouldPull = true
				} else if options.FixByDefault {
					fmt.Printf("Updating secret '%s' to correct format (--fix flag is enabled)\n", relativeFilename)
					shouldPull = true
				} else if !vars.IsTTY {
					fmt.Printf("Secret '%s' has the same data but in a different format, use the --fix flag to fix it\n", relativeFilename)
				} else {
					fmt.Printf("Secret '%s' has the same data but in a different format, do you want to fix it?", relativeFilename)
					shouldPull = util.CliQuestionYesNoDefault("Fix format?", true)
				}

				if shouldPull {
					err := project.pullSecret(secret, fetchedSecret, &fileState)
					if err != nil {
						return err
					}

					fmt.Println("    done")
				} else {
					fmt.Println("    skipped")
				}
			} else {
				// Local file format is correct

				if hasPrevState {
					if prevState.RemoteVersion != fileState.RemoteVersion {
						fmt.Printf("info: remote version for '%s' changed but is already in sync\n", relativeFilename)
					}

					if prevState.LocalHash != fileState.LocalHash {
						fmt.Printf("info: local secret '%s' contents changed but is already in sync\n", relativeFilename)
					}
				}

				fmt.Printf("Secret '%s' is already up to date\n", relativeFilename)
			}
		} else if fileState.LocalHash == "" && fileState.formatError != nil {
			// File exists but couldn't be parsed

			shouldPull := false

			if vars.IsCICD {
				fmt.Printf("Overwriting secret that failed parsing '%s' (--cicd flag is enabled)\n", relativeFilename)
				shouldPull = true
			} else if options.FixByDefault {
				fmt.Printf("Overwriting secret that failed parsing '%s' (--fix flag is enabled)\n", relativeFilename)
				shouldPull = true
			} else if !vars.IsTTY {
				fmt.Printf("Failed to parse secret '%s', use the --fix flag to fix it\n", relativeFilename)
			} else {
				fmt.Printf("Failed to parse secret '%s', do you want to fix it?", relativeFilename)
				shouldPull = util.CliQuestionYesNoDefault("Overwrite file?", true)
			}

			if shouldPull {
				err := project.pullSecret(secret, fetchedSecret, &fileState)
				if err != nil {
					return err
				}

				fmt.Println("    done")
			} else {
				fmt.Println("    skipped")
			}
		} else {
			// File exists but doesn't match remote secret

			if !hasPrevState {
				// File exists, doesn't match remote secret, and isn't in lockfile

				shouldPush := false
				shouldPull := false

				if vars.IsCICD {
					fmt.Printf("Overwriting new secret file '%s' with remote copy (--cicd flag is enabled)\n", relativeFilename)
					shouldPull = true
				} else if options.PullByDefault {
					fmt.Printf("Overwiting new secret file '%s' with remote copy (--pull flag is enabled)\n", relativeFilename)
					shouldPull = true
				} else if options.PushByDefault {
					fmt.Printf("Overwriting new remote secret with local copy '%s' (--push flag is enabled)\n", relativeFilename)
					shouldPush = true
				} else if !vars.IsTTY {
					fmt.Printf("New secret file '%s' does not match remote copy\n", relativeFilename)
				} else {
					fmt.Printf("New secret file '%s' does not match remote copy, do you want to pull, push, or leave it as is?\n", relativeFilename)
					shouldPush, shouldPull = cliQuestionPushPull()
				}

				if shouldPush {
					err = project.pushSecret(fetchedSecret, &fileState)
					if err != nil {
						return err
					}

					fmt.Println("    pushed")
				} else if shouldPull {
					err := project.pullSecret(secret, fetchedSecret, &fileState)
					if err != nil {
						return err
					}

					fmt.Println("    pulled")
				} else {
					fmt.Println("    skipped")
				}
			} else {
				// File exists, doesn't match remote secret, and has record in lockfile

				if fetchedSecret.Version() != prevState.RemoteVersion {
					// Remote version changed

					if fileState.LocalHash != prevState.LocalHash {
						// Local version changed too

						shouldPush := false
						shouldPull := false

						if vars.IsCICD {
							fmt.Printf("Overwriting modified secret file '%s' with remote copy (--cicd flag is enabled)\n", relativeFilename)
							shouldPull = true
						} else if options.PullByDefault {
							fmt.Printf("Overwiting modified secret file '%s' with remote copy (--pull flag is enabled)\n", relativeFilename)
							shouldPull = true
						} else if options.PushByDefault {
							fmt.Printf("Overwriting remote secret with modified local copy '%s' (--push flag is enabled)\n", relativeFilename)
							shouldPush = true
						} else if !vars.IsTTY {
							fmt.Printf("Modified secret file '%s' does not match modified remote copy\n", relativeFilename)
						} else {
							fmt.Printf("Modified secret file '%s' does not match modified remote copy, do you want to pull, push, or leave it as is?\n", relativeFilename)
							shouldPush, shouldPull = cliQuestionPushPull()
						}

						if shouldPush {
							err = project.pushSecret(fetchedSecret, &fileState)
							if err != nil {
								return err
							}

							fmt.Println("    pushed")
						} else if shouldPull {
							err := project.pullSecret(secret, fetchedSecret, &fileState)
							if err != nil {
								return err
							}

							fmt.Println("    pulled")
						} else {
							fmt.Println("    skipped")
						}
					} else {
						// Only the remote version changed

						fmt.Printf("Pulling new version of secret '%s'\n", relativeFilename)

						err := project.pullSecret(secret, fetchedSecret, &fileState)
						if err != nil {
							return err
						}

						fmt.Println("    done")
					}
				} else {
					// Remote version didn't change

					if fileState.LocalHash != prevState.LocalHash {
						// Only the local version changed

						fmt.Printf("Pushing new version of secret '%s'\n", relativeFilename)

						err := project.pushSecret(fetchedSecret, &fileState)
						if err != nil {
							return err
						}

						fmt.Println("    done")
					} else {
						// Neither version changed, lockfile is corrupt

						shouldPush := false
						shouldPull := false

						if vars.IsCICD {
							fmt.Printf("Lockfile is corrupt, overwriting secret file '%s' with remote copy (--cicd flag is enabled)\n", relativeFilename)
							shouldPull = true
						} else if !vars.IsTTY {
							fmt.Printf("Lockfile is corrupt, secret file '%s' does not match remote copy but neither are modified\n", relativeFilename)
						} else {
							fmt.Printf("Lockfile is corrupt, secret file '%s' does not match remote copy, do you want to pull, push, or leave it as is?\n", relativeFilename)
							shouldPush, shouldPull = cliQuestionPushPull()
						}

						if shouldPush {
							err = project.pushSecret(fetchedSecret, &fileState)
							if err != nil {
								return err
							}

							fmt.Println("    pushed")
						} else if shouldPull {
							err := project.pullSecret(secret, fetchedSecret, &fileState)
							if err != nil {
								return err
							}

							fmt.Println("    pulled")
						} else {
							fmt.Println("    skipped")
						}
					}
				}
			}
		}

		project.currentState.Files[secret.File] = fileState
	}

	for _, secret := range excludedSecrets {
		correctedFilename := filepath.Join(project.path, secret.File)
		relativeFilename, err := filepath.Rel(vars.Workdir, correctedFilename)

		if err != nil {
			return err
		}

		_, err = os.Stat(correctedFilename)
		if err == nil {
			var shouldDeleteFile bool

			if vars.IsCICD {
				fmt.Printf("Deleting unreferenced secret file at '%s' (--cicd flag is enabled)\n", relativeFilename)
				shouldDeleteFile = true
			} else if !vars.IsTTY {
				fmt.Printf("Warning: unreferenced secret file at '%s'\n", relativeFilename)
				shouldDeleteFile = false
			} else {
				fmt.Printf("Unreferenced file at %s, would you like to remove it?\n", relativeFilename)
				shouldDeleteFile = util.CliQuestionYesNoDefault("Delete file?", true)
			}

			if shouldDeleteFile {
				err := os.Remove(correctedFilename)
				if err != nil {
					return err
				}

				fmt.Println("    deleted")
			} else {
				fmt.Println("    skipped")
			}
		}
	}

	for filename := range project.lastState.Files {
		if _, exists := project.currentState.Files[filename]; !exists {
			correctedFilename := filepath.Join(project.path, filename)
			relativeFilename, err := filepath.Rel(vars.Workdir, correctedFilename)

			if err != nil {
				return err
			}

			_, err = os.Stat(correctedFilename)
			if err == nil {
				var shouldDeleteFile bool

				if vars.IsCICD {
					fmt.Printf("Deleting removed secret file at '%s' (--cicd flag is enabled)\n", relativeFilename)
					shouldDeleteFile = true
				} else if !vars.IsTTY {
					fmt.Printf("Warning: removed secret file at '%s'\n", relativeFilename)
					shouldDeleteFile = false
				} else {
					fmt.Printf("Removed file at %s, would you like to remove it?\n", relativeFilename)
					shouldDeleteFile = util.CliQuestionYesNoDefault("Delete file?", true)
				}

				if shouldDeleteFile {
					err := os.Remove(correctedFilename)
					if err != nil {
						return err
					}

					fmt.Println("    deleted")
				} else {
					fmt.Println("    skipped")
				}
			}
		}
	}

	err = project.saveCurrentState()
	if err != nil {
		return err
	}

	return nil
}

func (project *Project) pullSecret(
	secret SecretConfig,
	fetchedSecret types.FetchedSecret,
	fileState *LockedFile,
) error {
	correctedFilename := filepath.Join(project.path, secret.File)

	hash, err := hashValue(fetchedSecret.Value())
	if err != nil {
		return err
	}

	formattedData, err := util.FormatData(
		fetchedSecret.Value(),
		fetchedSecret.Format(),
	)
	if err != nil {
		return err
	}

	fileState.LocalHash = hash
	fileState.LocalFormat = util.FormatToName(fetchedSecret.Format())

	err = ioutil.WriteFile(
		correctedFilename,
		[]byte(formattedData),
		0774,
	)
	if err != nil {
		return err
	}

	return nil
}

func (project *Project) pushSecret(
	fetchedSecret types.FetchedSecret,
	fileState *LockedFile,
) error {
	newVersion, err := fetchedSecret.UploadNew(fileState.data)
	if err != nil {
		return err
	}

	fileState.RemoteVersion = newVersion

	return nil
}

func cliQuestionPushPull() (bool, bool) {
	for {
		answer := util.CliQuestion("Push (u), pull (d), or skip (n)?")

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "u", "push":
			return true, false
		case "d", "pull":
			return false, true
		case "n", "skip":
			return false, false
		default:
			fmt.Println("Please enter one of 'u', 'd', 'n', 'push', 'pull', or 'skip'")
		}
	}
}
