package main

import (
	_ "embed"
	"strings"
)

// Shell initialization script templates
// These are embedded at compile time using Go's embed directive

//go:embed resources/shell-init/bash.sh
var bashTemplate string

//go:embed resources/shell-init/zsh.sh
var zshTemplate string

//go:embed resources/shell-init/fish.sh
var fishTemplate string

// getShellTemplate returns the appropriate shell template for the given shell type
func getShellTemplate(shellType string) string {
	switch shellType {
	case "zsh":
		return zshTemplate
	case "fish":
		return fishTemplate
	default: // bash
		return bashTemplate
	}
}

// renderShellTemplate replaces template variables with actual values
func renderShellTemplate(template, cmdName string) string {
	// Replace all occurrences of {{CMD_NAME}} with the actual command name
	return strings.ReplaceAll(template, "{{CMD_NAME}}", cmdName)
}