package cmd

import (
	"fmt"
	"os"

	"github.com/madwire-media/secrets-cli/engines/vault"
	"github.com/madwire-media/secrets-cli/project"
	"github.com/madwire-media/secrets-cli/util"
	"github.com/madwire-media/secrets-cli/vars"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file path>",
	Short: "Add a file to the secrets.yaml",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		project, err := project.OpenProject()
		if err != nil {
			fmt.Println("Error opening project:", err)
			os.Exit(1)
			return
		}

		for _, secret := range project.Config.Secrets {
			if secret.File == file {
				fmt.Println("File already exists in secrets.yaml")
				os.Exit(1)
				return
			}
		}

		err = addSecret(file, project)
		if err != nil {
			fmt.Println("Error adding secret:", err)
			os.Exit(1)
			return
		}

		err = project.Save()
		if err != nil {
			fmt.Println("Error saving project:", err)
			os.Exit(1)
			return
		}

		fmt.Println("secrets.yaml updated")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func addSecret(file string, openProject *project.Project) error {
	class := util.CliQuestion("Secret class (optional)")

	err := util.LoadUserAuth()
	if err != nil {
		return err
	}

	var vaultConfig vault.SecretConfig

	choices := []string{}
	for host := range *vars.UserAuth.Vault {
		choices = append(choices, host)
	}
	choices = append(choices, "(other)")

	vaultHostChoice, err := util.CliChoice("Vault host", choices)
	if err != nil {
		return err
	}

	var host string

	if vaultHostChoice == len(choices)-1 {
		// Other choice
		host = util.CliQuestion("Custom Vault host (i.e. example.com:8080)")
	} else {
		host = choices[vaultHostChoice]
	}

	// TODO: use Vault /sys/mounts API to list available secrets engines?

	path := util.CliQuestion("Path to secret in Vault (i.e. secrets-engine/path/to/secret)")

	vaultConfig.URL = "https://" + host + "/" + path

	choices = []string{
		"From data (map a portion of Vault secret as JSON or YAML)",
		"From text (map a string value within a Vault secret)",
	}

	vaultMappingChoice, err := util.CliChoice("How to map the Vault secret to a local file", choices)

	if vaultMappingChoice == 0 {
		// fromData

		choices = []string{
			"JSON",
			"YAML",
		}

		vaultDataFormatChoice, err := util.CliChoice("Local file format", choices)
		if err != nil {
			return err
		}

		var vaultDataFormat string

		switch vaultDataFormatChoice {
		case 0:
			vaultDataFormat = "json"

		case 1:
			vaultDataFormat = "yaml"
		}

		var path *[]interface{}
		rawPath := util.CliQuestion("Path to data within Vault secret (optional)")

		if rawPath != "" {
			parsedPath := parsePath(rawPath)
			path = &parsedPath
		}

		vaultConfig.Mapping.FromData = &vault.FromDataMapping{
			Format: vaultDataFormat,
			Path:   path,
		}
	} else {
		// fromText

		rawPath := util.CliQuestion("Path to data within Vault secret")

		vaultConfig.Mapping.FromText = &vault.FromTextMapping{
			Path: parsePath(rawPath),
		}
	}

	var usedClass *string

	if class != "" {
		usedClass = &class
	}

	openProject.Config.Secrets = append(
		openProject.Config.Secrets,
		project.SecretConfig{
			File:  file,
			Class: usedClass,
			Vault: &vaultConfig,
		},
	)

	return nil
}

func parsePath(rawPath string) []interface{} {
	output := []interface{}{}
	var scratch string

	lastWasBackslash := false

	// TODO: support array indexes

	for _, char := range rawPath {
		switch char {
		case '\\':
			if lastWasBackslash {
				scratch += "\\"
				lastWasBackslash = false
			} else {
				lastWasBackslash = true
			}

		case '.':
			if lastWasBackslash {
				scratch += "."
				lastWasBackslash = false
			} else {
				output = append(output, scratch)
				scratch = ""
			}

		default:
			if lastWasBackslash {
				scratch += "\\"
				lastWasBackslash = false
			}

			scratch += string(char)
		}
	}

	return output
}
