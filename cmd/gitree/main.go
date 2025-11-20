package main

import (
	"fmt"
	"os"
)

var (
	// Version information (injected at build time via ldflags).
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
