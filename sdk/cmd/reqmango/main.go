// Command reqmango is the reqmango daily-ops CLI.
package main

import (
	"os"

	"github.com/reqmango/tools/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
