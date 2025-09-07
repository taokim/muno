package adapters

import (
	"fmt"
	"io"
)

// RealOutput implements interfaces.Output using io.Writer
type RealOutput struct {
	stdout io.Writer
	stderr io.Writer
}

// NewRealOutput creates a new real output implementation
func NewRealOutput(stdout, stderr io.Writer) *RealOutput {
	return &RealOutput{
		stdout: stdout,
		stderr: stderr,
	}
}

// Write implements io.Writer
func (o *RealOutput) Write(p []byte) (n int, err error) {
	return o.stdout.Write(p)
}

// Print methods
func (o *RealOutput) Print(a ...interface{}) (n int, err error) {
	if o.stdout == nil {
		return 0, nil
	}
	return fmt.Fprint(o.stdout, a...)
}

func (o *RealOutput) Printf(format string, a ...interface{}) (n int, err error) {
	if o.stdout == nil {
		return 0, nil
	}
	return fmt.Fprintf(o.stdout, format, a...)
}

func (o *RealOutput) Println(a ...interface{}) (n int, err error) {
	if o.stdout == nil {
		return 0, nil
	}
	return fmt.Fprintln(o.stdout, a...)
}

// Error output
func (o *RealOutput) Error(a ...interface{}) (n int, err error) {
	if o.stderr == nil {
		return 0, nil
	}
	return fmt.Fprint(o.stderr, a...)
}

func (o *RealOutput) Errorf(format string, a ...interface{}) (n int, err error) {
	if o.stderr == nil {
		return 0, nil
	}
	return fmt.Fprintf(o.stderr, format, a...)
}

func (o *RealOutput) Errorln(a ...interface{}) (n int, err error) {
	if o.stderr == nil {
		return 0, nil
	}
	return fmt.Fprintln(o.stderr, a...)
}

// Colored output (using simple markers for now, can be enhanced with color libraries)
func (o *RealOutput) Success(message string) {
	if o.stdout != nil {
		fmt.Fprintf(o.stdout, "✅ %s\n", message)
	}
}

func (o *RealOutput) Info(message string) {
	if o.stdout != nil {
		fmt.Fprintf(o.stdout, "ℹ️ %s\n", message)
	}
}

func (o *RealOutput) Warning(message string) {
	if o.stdout != nil {
		fmt.Fprintf(o.stdout, "⚠️ %s\n", message)
	}
}

func (o *RealOutput) Danger(message string) {
	if o.stderr != nil {
		fmt.Fprintf(o.stderr, "🚨 %s\n", message)
	}
}

// Formatting (basic implementation, can be enhanced with ANSI codes)
func (o *RealOutput) Bold(text string) string {
	// In a real implementation, this would use ANSI escape codes
	return text
}

func (o *RealOutput) Italic(text string) string {
	// In a real implementation, this would use ANSI escape codes
	return text
}

func (o *RealOutput) Underline(text string) string {
	// In a real implementation, this would use ANSI escape codes
	return text
}

func (o *RealOutput) Color(text string, color string) string {
	// In a real implementation, this would use ANSI escape codes based on color
	return text
}