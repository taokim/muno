package adapters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealOutput_Print(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "Simple message",
			message:        "Hello, World!",
			expectedStdout: "Hello, World!",
			expectedStderr: "",
		},
		{
			name:           "Empty message",
			message:        "",
			expectedStdout: "",
			expectedStderr: "",
		},
		{
			name:           "Message with newline",
			message:        "Line 1\nLine 2",
			expectedStdout: "Line 1\nLine 2",
			expectedStderr: "",
		},
		{
			name:           "Message with special characters",
			message:        "Special: !@#$%^&*()",
			expectedStdout: "Special: !@#$%^&*()",
			expectedStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			output := NewRealOutput(&stdout, &stderr)

			output.Print(tt.message)

			assert.Equal(t, tt.expectedStdout, stdout.String())
			assert.Equal(t, tt.expectedStderr, stderr.String())
		})
	}
}

func TestRealOutput_Println(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "Simple message",
			message:        "Hello, World!",
			expectedStdout: "Hello, World!\n",
			expectedStderr: "",
		},
		{
			name:           "Empty message",
			message:        "",
			expectedStdout: "\n",
			expectedStderr: "",
		},
		{
			name:           "Message with existing newline",
			message:        "Already has newline\n",
			expectedStdout: "Already has newline\n\n",
			expectedStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			output := NewRealOutput(&stdout, &stderr)

			output.Println(tt.message)

			assert.Equal(t, tt.expectedStdout, stdout.String())
			assert.Equal(t, tt.expectedStderr, stderr.String())
		})
	}
}

func TestRealOutput_Printf(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		args           []interface{}
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "Simple format",
			format:         "Hello, %s!",
			args:           []interface{}{"World"},
			expectedStdout: "Hello, World!",
			expectedStderr: "",
		},
		{
			name:           "Multiple arguments",
			format:         "%s is %d years old",
			args:           []interface{}{"Alice", 30},
			expectedStdout: "Alice is 30 years old",
			expectedStderr: "",
		},
		{
			name:           "No arguments",
			format:         "No formatting",
			args:           []interface{}{},
			expectedStdout: "No formatting",
			expectedStderr: "",
		},
		{
			name:           "Various types",
			format:         "String: %s, Int: %d, Float: %.2f, Bool: %t",
			args:           []interface{}{"test", 42, 3.14159, true},
			expectedStdout: "String: test, Int: 42, Float: 3.14, Bool: true",
			expectedStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			output := NewRealOutput(&stdout, &stderr)

			output.Printf(tt.format, tt.args...)

			assert.Equal(t, tt.expectedStdout, stdout.String())
			assert.Equal(t, tt.expectedStderr, stderr.String())
		})
	}
}

func TestRealOutput_Error(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "Simple error",
			message:        "Error: something went wrong",
			expectedStdout: "",
			expectedStderr: "Error: something went wrong",
		},
		{
			name:           "Empty error",
			message:        "",
			expectedStdout: "",
			expectedStderr: "",
		},
		{
			name:           "Multi-line error",
			message:        "Error line 1\nError line 2",
			expectedStdout: "",
			expectedStderr: "Error line 1\nError line 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			output := NewRealOutput(&stdout, &stderr)

			output.Error(tt.message)

			assert.Equal(t, tt.expectedStdout, stdout.String())
			assert.Equal(t, tt.expectedStderr, stderr.String())
		})
	}
}

func TestRealOutput_Errorln(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "Simple error",
			message:        "Error: something went wrong",
			expectedStdout: "",
			expectedStderr: "Error: something went wrong\n",
		},
		{
			name:           "Empty error",
			message:        "",
			expectedStdout: "",
			expectedStderr: "\n",
		},
		{
			name:           "Error with existing newline",
			message:        "Already has newline\n",
			expectedStdout: "",
			expectedStderr: "Already has newline\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			output := NewRealOutput(&stdout, &stderr)

			output.Errorln(tt.message)

			assert.Equal(t, tt.expectedStdout, stdout.String())
			assert.Equal(t, tt.expectedStderr, stderr.String())
		})
	}
}

func TestRealOutput_Errorf(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		args           []interface{}
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "Simple error format",
			format:         "Error: %s",
			args:           []interface{}{"file not found"},
			expectedStdout: "",
			expectedStderr: "Error: file not found",
		},
		{
			name:           "Error with code",
			format:         "Error %d: %s",
			args:           []interface{}{404, "not found"},
			expectedStdout: "",
			expectedStderr: "Error 404: not found",
		},
		{
			name:           "No arguments",
			format:         "Generic error",
			args:           []interface{}{},
			expectedStdout: "",
			expectedStderr: "Generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			output := NewRealOutput(&stdout, &stderr)

			output.Errorf(tt.format, tt.args...)

			assert.Equal(t, tt.expectedStdout, stdout.String())
			assert.Equal(t, tt.expectedStderr, stderr.String())
		})
	}
}

func TestRealOutput_Mixed(t *testing.T) {
	t.Run("Mixed stdout and stderr", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		output := NewRealOutput(&stdout, &stderr)

		// Write to both stdout and stderr
		output.Print("stdout: ")
		output.Error("stderr: ")
		output.Println("line 1")
		output.Errorln("error 1")
		output.Printf("formatted %d", 42)
		output.Errorf("error %d", 99)

		// Check stdout
		expectedStdout := "stdout: line 1\nformatted 42"
		assert.Equal(t, expectedStdout, stdout.String())

		// Check stderr
		expectedStderr := "stderr: error 1\nerror 99"
		assert.Equal(t, expectedStderr, stderr.String())
	})

	t.Run("Sequential operations", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		output := NewRealOutput(&stdout, &stderr)

		// Perform sequential operations
		output.Println("First line")
		output.Println("Second line")
		output.Errorln("First error")
		output.Print("Third")
		output.Print(" line")
		output.Println("")
		output.Error("Second")
		output.Errorln(" error")

		// Check stdout
		expectedStdout := "First line\nSecond line\nThird line\n"
		assert.Equal(t, expectedStdout, stdout.String())

		// Check stderr
		expectedStderr := "First error\nSecond error\n"
		assert.Equal(t, expectedStderr, stderr.String())
	})
}

func TestRealOutput_LargeOutput(t *testing.T) {
	t.Run("Large stdout output", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		output := NewRealOutput(&stdout, &stderr)

		// Generate large output
		var expected strings.Builder
		for i := 0; i < 1000; i++ {
			msg := "Line " + string(rune(i))
			output.Println(msg)
			expected.WriteString(msg)
			expected.WriteString("\n")
		}

		assert.Equal(t, expected.String(), stdout.String())
		assert.Empty(t, stderr.String())
	})

	t.Run("Large stderr output", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		output := NewRealOutput(&stdout, &stderr)

		// Generate large error output
		var expected strings.Builder
		for i := 0; i < 1000; i++ {
			msg := "Error " + string(rune(i))
			output.Errorln(msg)
			expected.WriteString(msg)
			expected.WriteString("\n")
		}

		assert.Empty(t, stdout.String())
		assert.Equal(t, expected.String(), stderr.String())
	})
}

func TestRealOutput_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Unicode characters",
			message: "Hello 世界 🌍",
		},
		{
			name:    "Tab characters",
			message: "Column1\tColumn2\tColumn3",
		},
		{
			name:    "Carriage return",
			message: "Line1\rLine2",
		},
		{
			name:    "Null bytes",
			message: "Before\x00After",
		},
		{
			name:    "ANSI escape codes",
			message: "\033[31mRed Text\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			output := NewRealOutput(&stdout, &stderr)

			// Test stdout
			output.Print(tt.message)
			assert.Equal(t, tt.message, stdout.String())

			// Reset buffers
			stdout.Reset()
			stderr.Reset()

			// Test stderr
			output.Error(tt.message)
			assert.Equal(t, tt.message, stderr.String())
		})
	}
}

func TestRealOutput_NilWriters(t *testing.T) {
	t.Run("Nil stdout", func(t *testing.T) {
		var stderr bytes.Buffer
		output := NewRealOutput(nil, &stderr)

		// These should not panic
		output.Print("test")
		output.Println("test")
		output.Printf("test %d", 42)

		// Stderr should still work
		output.Error("error")
		assert.Equal(t, "error", stderr.String())
	})

	t.Run("Nil stderr", func(t *testing.T) {
		var stdout bytes.Buffer
		output := NewRealOutput(&stdout, nil)

		// These should not panic
		output.Error("error")
		output.Errorln("error")
		output.Errorf("error %d", 42)

		// Stdout should still work
		output.Print("test")
		assert.Equal(t, "test", stdout.String())
	})

	t.Run("Both nil", func(t *testing.T) {
		output := NewRealOutput(nil, nil)

		// None of these should panic
		output.Print("test")
		output.Println("test")
		output.Printf("test %d", 42)
		output.Error("error")
		output.Errorln("error")
		output.Errorf("error %d", 42)
	})
}