package adapters

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUIAdapter(t *testing.T) {
	ui := NewUIAdapter()
	assert.NotNil(t, ui)
	
	// Check it returns the correct interface type
	_, ok := ui.(*UIAdapter)
	assert.True(t, ok)
}

func TestUIAdapter_Info(t *testing.T) {
	ui := &UIAdapter{}
	
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	ui.Info("test info message")
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	assert.Contains(t, output, "ℹ️  test info message")
}

func TestUIAdapter_Success(t *testing.T) {
	ui := &UIAdapter{}
	
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	ui.Success("operation successful")
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	assert.Contains(t, output, "✅ operation successful")
}

func TestUIAdapter_Warning(t *testing.T) {
	ui := &UIAdapter{}
	
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	ui.Warning("warning message")
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	assert.Contains(t, output, "⚠️  warning message")
}

func TestUIAdapter_Error(t *testing.T) {
	ui := &UIAdapter{}
	
	// Capture stdout (Error prints to stdout, not stderr)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	ui.Error("error message")
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	assert.Contains(t, output, "❌ error message")
}

func TestUIAdapter_Debug(t *testing.T) {
	// Test with debug disabled (default)
	ui := &UIAdapter{debug: false}
	
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	ui.Debug("debug message")
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	// Should not output when debug is false
	assert.Empty(t, output)
	
	// Test with debug enabled
	ui = &UIAdapter{debug: true}
	
	old = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	
	ui.Debug("debug message")
	
	w.Close()
	os.Stdout = old
	
	buf.Reset()
	io.Copy(&buf, r)
	output = buf.String()
	
	assert.Contains(t, output, "[DEBUG] debug message")
}

func TestUIAdapter_Confirm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Yes", "y\n", true},
		{"YES", "Y\n", true},
		{"yes", "yes\n", true},
		{"No", "n\n", false},
		{"NO", "N\n", false},
		{"no", "no\n", false},
		{"Empty", "\n", false},
		{"Other", "maybe\n", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := &UIAdapter{
				reader: bufio.NewReader(strings.NewReader(tt.input)),
			}
			
			// Capture stdout to suppress prompt
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			
			result, err := ui.Confirm("Continue?")
			assert.NoError(t, err)
			
			w.Close()
			os.Stdout = old
			r.Close()
			
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUIAdapter_Select(t *testing.T) {
	// Test valid selection
	ui := &UIAdapter{
		reader: bufio.NewReader(strings.NewReader("1\n")),
	}
	
	// Capture stdout to suppress prompt
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	options := []string{"option1", "option2", "option3"}
	result, err := ui.Select("Choose one:", options)
	
	w.Close()
	os.Stdout = old
	r.Close()
	
	assert.NoError(t, err)
	assert.Equal(t, "option1", result)
	
	// Test invalid selection
	ui = &UIAdapter{
		reader: bufio.NewReader(strings.NewReader("5\n")),
	}
	
	old = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	
	result, err = ui.Select("Choose:", options)
	
	w.Close()
	os.Stdout = old
	r.Close()
	
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestUIAdapter_MultiSelect(t *testing.T) {
	// Test valid multi-selection
	ui := &UIAdapter{
		reader: bufio.NewReader(strings.NewReader("1,3\n")),
	}
	
	// Capture stdout to suppress prompt
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	options := []string{"option1", "option2", "option3"}
	results, err := ui.MultiSelect("Choose multiple:", options)
	
	w.Close()
	os.Stdout = old
	r.Close()
	
	assert.NoError(t, err)
	assert.Equal(t, []string{"option1", "option3"}, results)
	
	// Test empty input
	ui = &UIAdapter{
		reader: bufio.NewReader(strings.NewReader("\n")),
	}
	
	old = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	
	results, err = ui.MultiSelect("Choose:", options)
	
	w.Close()
	os.Stdout = old
	r.Close()
	
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestUIAdapter_Progress(t *testing.T) {
	ui := &UIAdapter{}
	
	// Test creating progress
	progress := ui.Progress("Processing...")
	assert.NotNil(t, progress)
	
	// Test progress methods
	cp, ok := progress.(*consoleProgress)
	assert.True(t, ok)
	
	// Capture output for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	cp.Start()
	cp.Update(50, 100)
	cp.SetMessage("Halfway done")
	cp.Finish()
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	// Should contain progress indicators
	assert.Contains(t, output, "Processing...")
	assert.Contains(t, output, "50%")
	assert.Contains(t, output, "Halfway done")
	assert.Contains(t, output, "done")
}

func TestConsoleProgress_Error(t *testing.T) {
	cp := &consoleProgress{
		message: "Test progress",
	}
	
	// Capture stdout (Error prints to stdout)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	err := fmt.Errorf("test error")
	cp.Error(err)
	
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	assert.Contains(t, output, "error")
	assert.Contains(t, output, "test error")
}