package main

import (
	"fmt"
	"os"
	"pantui/cmd"
)

// Version information - set by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info for cmd package
	cmd.SetVersionInfo(version, commit, date)
	
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
