package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	app := NewApp()
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}