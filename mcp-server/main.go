package main

import (
	"fmt"
	"os"
)

func main() {
	cfg := LoadConfig()

	if cfg.APIToken == "" {
		fmt.Fprintln(os.Stderr, "ERROR: reqmango_API_TOKEN environment variable is required.")
		fmt.Fprintln(os.Stderr, "Set it to a valid reqmango JWT Bearer token.")
		os.Exit(1)
	}

	server := NewServer(cfg)

	if err := server.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
