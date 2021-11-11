package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/madwire-media/secrets-cli/project"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [classes]",
	Short: "Sync secrets in local project",
	Long: `Fetch the latest versions of secrets and update the local copies if necessary
Any class settings you choose will be saved into a local class file for future
use. Secrets without an assigned class will always be synced.`,
	Example: `    [classes]:
        (empty)    uses only classes saved in local class file
        +all       adds all classes (overwrites class file settings)
        +foo,+bar  adds the 'foo' and 'bar' classes
        ,-foo      removes the 'foo' class
        +all,-foo  adds all classes except 'foo' (overwrites class file settings)
        ,-all      resets all classes (overwrites class file settings)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		openProject, err := project.OpenProject()
		if err != nil {
			fmt.Println("Error opening project:", err)
			os.Exit(1)
			return
		}

		pullOnly, _ := cmd.Flags().GetBool("pull")
		pushOnly, _ := cmd.Flags().GetBool("push")
		fixByDefault, _ := cmd.Flags().GetBool("fix")

		options := project.SyncOptions{
			PullOnly:     pullOnly,
			PushOnly:     pushOnly,
			FixByDefault: fixByDefault,
			Classes: project.ClassUpdate{
				FilterOptions: project.FilterOptions{
					Add:      []string{},
					Subtract: []string{},
				},
			},
		}

		if len(args) > 0 {
			for _, class := range strings.Split(args[0], ",") {
				class = strings.TrimSpace(class)

				if class == "-all" {
					options.Classes.Reset = true
				} else if class == "+all" {
					options.Classes.DefaultAll = true
				} else if strings.HasPrefix(class, "+") {
					options.Classes.Add = append(options.Classes.Add, class[1:])
				} else if strings.HasPrefix(class, "-") {
					options.Classes.Subtract = append(options.Classes.Subtract, class[1:])
				} else if class != "" {
					panic("Unexpected class arg '" + class + "'")
				}
			}
		}

		err = openProject.Sync(options)
		if err != nil {
			fmt.Println("Error syncing secrets:", err)
			os.Exit(1)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().Bool("pull", false, "prefer pulling remote secrets during conflicts, and don't push local changes")
	syncCmd.Flags().Bool("push", false, "prefer pushing local changes during conflicts, and don't pull remote changes")
	syncCmd.Flags().Bool("fix", false, "fix issues with secrets by default")
}
