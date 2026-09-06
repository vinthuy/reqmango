// Command reqmango is the reqmango daily-ops CLI.
package main

import (
	"fmt"
	"os"

	"github.com/reqmango/tools/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		// SilenceErrors keeps cobra from printing; main owns the single
		// error line so hints like the 401 "run `reqmango auth login`"
		// message are always visible to the user.
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
