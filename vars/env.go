package vars

import (
	"os"

	"github.com/mattn/go-isatty"
)

var (
	// Workdir is the working directory this program was started with
	Workdir string

	// IsCICD is true when the --cicd flag is enabled or the CICD environment
	// variable is non-null. The environment variable check is here in this
	// file, but the CLI flag definition is in cmd/root.go
	IsCICD bool

	// IsTTY is true when the program is running in a TTY and theoretically is
	// being piloted by a user. If IsTTY is false then there should not be any
	// CLI prompts
	IsTTY bool = isatty.IsTerminal(os.Stdout.Fd())
)

func init() {
	var err error

	Workdir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	// Can also be overridden with --cicd CLI flag in cmd/root.go
	IsCICD = os.Getenv("CICD") != ""
}
