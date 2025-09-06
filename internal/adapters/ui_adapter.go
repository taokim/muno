package adapters

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	
	"github.com/taokim/muno/internal/interfaces"
	"golang.org/x/term"
)

// UIAdapter provides console-based user interaction
type UIAdapter struct {
	reader *bufio.Reader
	debug  bool
}

// NewUIAdapter creates a new UI adapter
func NewUIAdapter() interfaces.UIProvider {
	return &UIAdapter{
		reader: bufio.NewReader(os.Stdin),
		debug:  os.Getenv("DEBUG") == "true",
	}
}

// Prompt prompts the user for input
func (u *UIAdapter) Prompt(message string) (string, error) {
	fmt.Print(message + " ")
	input, err := u.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// PromptPassword prompts for a password (hidden input)
func (u *UIAdapter) PromptPassword(message string) (string, error) {
	fmt.Print(message + " ")
	
	// Read password without echoing
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // New line after password input
	
	return string(password), nil
}

// Confirm asks for yes/no confirmation
func (u *UIAdapter) Confirm(message string) (bool, error) {
	response, err := u.Prompt(message + " (y/n)")
	if err != nil {
		return false, err
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// Select presents options for selection
func (u *UIAdapter) Select(message string, options []string) (string, error) {
	fmt.Println(message)
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}
	
	response, err := u.Prompt("Enter choice (number):")
	if err != nil {
		return "", err
	}
	
	// Parse selection
	var choice int
	_, err = fmt.Sscanf(response, "%d", &choice)
	if err != nil || choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid selection")
	}
	
	return options[choice-1], nil
}

// MultiSelect allows multiple selections
func (u *UIAdapter) MultiSelect(message string, options []string) ([]string, error) {
	fmt.Println(message)
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}
	
	response, err := u.Prompt("Enter choices (comma-separated numbers):")
	if err != nil {
		return nil, err
	}
	
	// Parse selections
	parts := strings.Split(response, ",")
	var selected []string
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		var choice int
		_, err = fmt.Sscanf(part, "%d", &choice)
		if err != nil || choice < 1 || choice > len(options) {
			continue
		}
		selected = append(selected, options[choice-1])
	}
	
	return selected, nil
}

// Progress creates a progress reporter
func (u *UIAdapter) Progress(message string) interfaces.ProgressReporter {
	return &consoleProgress{
		message: message,
	}
}

// Info displays an info message
func (u *UIAdapter) Info(message string) {
	fmt.Printf("%s\n", message)
}

// Success displays a success message
func (u *UIAdapter) Success(message string) {
	fmt.Printf("✅ %s\n", message)
}

// Warning displays a warning message
func (u *UIAdapter) Warning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// Error displays an error message
func (u *UIAdapter) Error(message string) {
	fmt.Printf("❌ %s\n", message)
}

// Debug displays a debug message
func (u *UIAdapter) Debug(message string) {
	if u.debug {
		fmt.Printf("🔍 [DEBUG] %s\n", message)
	}
}

// consoleProgress implements ProgressReporter for console output
type consoleProgress struct {
	message string
	current int
	total   int
}

// Start starts the progress indicator
func (p *consoleProgress) Start() {
	fmt.Printf("%s...\n", p.message)
}

// Update updates the progress
func (p *consoleProgress) Update(current, total int) {
	p.current = current
	p.total = total
	
	if total > 0 {
		percent := (current * 100) / total
		fmt.Printf("\r%s: %d/%d (%d%%)", p.message, current, total, percent)
	} else {
		fmt.Printf("\r%s: %d", p.message, current)
	}
}

// SetMessage updates the progress message
func (p *consoleProgress) SetMessage(message string) {
	p.message = message
	fmt.Printf("\r%s", message)
}

// Finish completes the progress
func (p *consoleProgress) Finish() {
	if p.total > 0 {
		fmt.Printf("\r%s: %d/%d (100%%)\n", p.message, p.total, p.total)
	} else {
		fmt.Printf("\r%s: done\n", p.message)
	}
}

// Error reports an error during progress
func (p *consoleProgress) Error(err error) {
	fmt.Printf("\r%s: error - %v\n", p.message, err)
}