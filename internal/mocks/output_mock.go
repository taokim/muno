package mocks

import (
	"bytes"
	"fmt"
	"sync"
)

// MockOutput implements interfaces.Output for testing
type MockOutput struct {
	mu     sync.RWMutex
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	
	// Call tracking
	Calls []string
	
	// Function overrides
	PrintFunc    func(a ...interface{}) (n int, err error)
	PrintfFunc   func(format string, a ...interface{}) (n int, err error)
	PrintlnFunc  func(a ...interface{}) (n int, err error)
	ErrorFunc    func(a ...interface{}) (n int, err error)
	ErrorfFunc   func(format string, a ...interface{}) (n int, err error)
	ErrorlnFunc  func(a ...interface{}) (n int, err error)
	SuccessFunc  func(message string)
	InfoFunc     func(message string)
	WarningFunc  func(message string)
	DangerFunc   func(message string)
	BoldFunc     func(text string) string
	ItalicFunc   func(text string) string
	UnderlineFunc func(text string) string
	ColorFunc    func(text string, color string) string
}

// NewMockOutput creates a new mock output
func NewMockOutput() *MockOutput {
	return &MockOutput{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		Calls:  []string{},
	}
}

// Write implements io.Writer
func (o *MockOutput) Write(p []byte) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	return o.stdout.Write(p)
}

// Print implements Output.Print
func (o *MockOutput) Print(a ...interface{}) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Print(%v)", a))
	
	if o.PrintFunc != nil {
		return o.PrintFunc(a...)
	}
	
	return fmt.Fprint(o.stdout, a...)
}

// Printf implements Output.Printf
func (o *MockOutput) Printf(format string, a ...interface{}) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Printf(%s, %v)", format, a))
	
	if o.PrintfFunc != nil {
		return o.PrintfFunc(format, a...)
	}
	
	return fmt.Fprintf(o.stdout, format, a...)
}

// Println implements Output.Println
func (o *MockOutput) Println(a ...interface{}) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Println(%v)", a))
	
	if o.PrintlnFunc != nil {
		return o.PrintlnFunc(a...)
	}
	
	return fmt.Fprintln(o.stdout, a...)
}

// Error implements Output.Error
func (o *MockOutput) Error(a ...interface{}) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Error(%v)", a))
	
	if o.ErrorFunc != nil {
		return o.ErrorFunc(a...)
	}
	
	return fmt.Fprint(o.stderr, a...)
}

// Errorf implements Output.Errorf
func (o *MockOutput) Errorf(format string, a ...interface{}) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Errorf(%s, %v)", format, a))
	
	if o.ErrorfFunc != nil {
		return o.ErrorfFunc(format, a...)
	}
	
	return fmt.Fprintf(o.stderr, format, a...)
}

// Errorln implements Output.Errorln
func (o *MockOutput) Errorln(a ...interface{}) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Errorln(%v)", a))
	
	if o.ErrorlnFunc != nil {
		return o.ErrorlnFunc(a...)
	}
	
	return fmt.Fprintln(o.stderr, a...)
}

// Success implements Output.Success
func (o *MockOutput) Success(message string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Success(%s)", message))
	
	if o.SuccessFunc != nil {
		o.SuccessFunc(message)
		return
	}
	
	fmt.Fprintf(o.stdout, "✅ %s\n", message)
}

// Info implements Output.Info
func (o *MockOutput) Info(message string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Info(%s)", message))
	
	if o.InfoFunc != nil {
		o.InfoFunc(message)
		return
	}
	
	fmt.Fprintf(o.stdout, "ℹ️ %s\n", message)
}

// Warning implements Output.Warning
func (o *MockOutput) Warning(message string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Warning(%s)", message))
	
	if o.WarningFunc != nil {
		o.WarningFunc(message)
		return
	}
	
	fmt.Fprintf(o.stdout, "⚠️ %s\n", message)
}

// Danger implements Output.Danger
func (o *MockOutput) Danger(message string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Danger(%s)", message))
	
	if o.DangerFunc != nil {
		o.DangerFunc(message)
		return
	}
	
	fmt.Fprintf(o.stderr, "🚨 %s\n", message)
}

// Bold implements Output.Bold
func (o *MockOutput) Bold(text string) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Bold(%s)", text))
	
	if o.BoldFunc != nil {
		return o.BoldFunc(text)
	}
	
	return fmt.Sprintf("**%s**", text)
}

// Italic implements Output.Italic
func (o *MockOutput) Italic(text string) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Italic(%s)", text))
	
	if o.ItalicFunc != nil {
		return o.ItalicFunc(text)
	}
	
	return fmt.Sprintf("*%s*", text)
}

// Underline implements Output.Underline
func (o *MockOutput) Underline(text string) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Underline(%s)", text))
	
	if o.UnderlineFunc != nil {
		return o.UnderlineFunc(text)
	}
	
	return fmt.Sprintf("_%s_", text)
}

// Color implements Output.Color
func (o *MockOutput) Color(text string, color string) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.Calls = append(o.Calls, fmt.Sprintf("Color(%s, %s)", text, color))
	
	if o.ColorFunc != nil {
		return o.ColorFunc(text, color)
	}
	
	return fmt.Sprintf("[%s]%s[/%s]", color, text, color)
}

// Helper methods for testing

// GetStdout returns the stdout buffer content
func (o *MockOutput) GetStdout() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	return o.stdout.String()
}

// GetStderr returns the stderr buffer content
func (o *MockOutput) GetStderr() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	return o.stderr.String()
}

// GetOutput returns both stdout and stderr
func (o *MockOutput) GetOutput() (stdout, stderr string) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	return o.stdout.String(), o.stderr.String()
}

// GetCalls returns all method calls
func (o *MockOutput) GetCalls() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	calls := make([]string, len(o.Calls))
	copy(calls, o.Calls)
	return calls
}

// Reset clears all buffers and calls
func (o *MockOutput) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.stdout.Reset()
	o.stderr.Reset()
	o.Calls = []string{}
}