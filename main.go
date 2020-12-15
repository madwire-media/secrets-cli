package main

import (
	"github.com/madwire-media/secrets-cli/cmd"
	"github.com/madwire-media/secrets-cli/util"
)

func main() {
	util.TryHandleSudo()

	cmd.Execute()
}
