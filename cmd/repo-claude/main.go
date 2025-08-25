package main

import (
	"fmt"
	"os"
)

// Build-time variables (set via ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
	gitBranch = "unknown"
)

func main() {
	app := NewApp()
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}