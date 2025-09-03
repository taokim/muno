package mocks

import (
	"fmt"
	"sync"
	
	"github.com/taokim/muno/internal/interfaces"
)

// MockUIProvider is a mock implementation of UIProvider
type MockUIProvider struct {
	mu         sync.RWMutex
	responses  map[string]string
	confirms   map[string]bool
	selections map[string]string
	calls      []string
	messages   []string
}

// NewMockUIProvider creates a new mock UI provider
func NewMockUIProvider() *MockUIProvider {
	return &MockUIProvider{
		responses:  make(map[string]string),
		confirms:   make(map[string]bool),
		selections: make(map[string]string),
		calls:      []string{},
		messages:   []string{},
	}
}

// Prompt prompts the user for input
func (m *MockUIProvider) Prompt(message string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Prompt(%s)", message))
	
	if response, ok := m.responses[message]; ok {
		return response, nil
	}
	
	return "default", nil
}

// PromptPassword prompts for a password
func (m *MockUIProvider) PromptPassword(message string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("PromptPassword(%s)", message))
	
	if response, ok := m.responses["password:"+message]; ok {
		return response, nil
	}
	
	return "password123", nil
}

// Confirm asks for yes/no confirmation
func (m *MockUIProvider) Confirm(message string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Confirm(%s)", message))
	
	if confirm, ok := m.confirms[message]; ok {
		return confirm, nil
	}
	
	// Check for default response
	if confirm, ok := m.confirms[""]; ok {
		return confirm, nil
	}
	
	return true, nil // Default to yes for testing
}

// Select presents options for selection
func (m *MockUIProvider) Select(message string, options []string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Select(%s, %v)", message, options))
	
	if selection, ok := m.selections[message]; ok {
		return selection, nil
	}
	
	if len(options) > 0 {
		return options[0], nil // Default to first option
	}
	
	return "", fmt.Errorf("no options provided")
}

// MultiSelect allows multiple selections
func (m *MockUIProvider) MultiSelect(message string, options []string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("MultiSelect(%s, %v)", message, options))
	
	if selection, ok := m.selections["multi:"+message]; ok {
		return []string{selection}, nil
	}
	
	if len(options) > 0 {
		return []string{options[0]}, nil // Default to first option
	}
	
	return []string{}, nil
}

// Progress creates a progress reporter
func (m *MockUIProvider) Progress(message string) interfaces.ProgressReporter {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Progress(%s)", message))
	
	return &mockProgress{
		message: message,
		ui:      m,
	}
}

// Info displays an info message
func (m *MockUIProvider) Info(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Info(%s)", message))
	m.messages = append(m.messages, "INFO: "+message)
}

// Success displays a success message
func (m *MockUIProvider) Success(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Success(%s)", message))
	m.messages = append(m.messages, "SUCCESS: "+message)
}

// Warning displays a warning message
func (m *MockUIProvider) Warning(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Warning(%s)", message))
	m.messages = append(m.messages, "WARNING: "+message)
}

// Error displays an error message
func (m *MockUIProvider) Error(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Error(%s)", message))
	m.messages = append(m.messages, "ERROR: "+message)
}

// Debug displays a debug message
func (m *MockUIProvider) Debug(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("Debug(%s)", message))
	m.messages = append(m.messages, "DEBUG: "+message)
}

// Mock helper methods

// SetResponse sets a mock response for prompts
func (m *MockUIProvider) SetResponse(prompt, response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.responses[prompt] = response
}

// SetConfirm sets a mock confirmation response
func (m *MockUIProvider) SetConfirm(message string, confirm bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.confirms[message] = confirm
}

// SetSelection sets a mock selection response
func (m *MockUIProvider) SetSelection(message, selection string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.selections[message] = selection
}

// GetCalls returns all method calls made
func (m *MockUIProvider) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetMessages returns all messages displayed
func (m *MockUIProvider) GetMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	messages := make([]string, len(m.messages))
	copy(messages, m.messages)
	return messages
}

// Reset resets the mock state
func (m *MockUIProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.responses = make(map[string]string)
	m.confirms = make(map[string]bool)
	m.selections = make(map[string]string)
	m.calls = []string{}
	m.messages = []string{}
}

// mockProgress implements ProgressReporter for testing
type mockProgress struct {
	message string
	ui      *MockUIProvider
	current int
	total   int
}

func (p *mockProgress) Start() {
	p.ui.mu.Lock()
	defer p.ui.mu.Unlock()
	
	p.ui.messages = append(p.ui.messages, fmt.Sprintf("PROGRESS_START: %s", p.message))
}

func (p *mockProgress) Update(current, total int) {
	p.ui.mu.Lock()
	defer p.ui.mu.Unlock()
	
	p.current = current
	p.total = total
	p.ui.messages = append(p.ui.messages, fmt.Sprintf("PROGRESS_UPDATE: %s [%d/%d]", p.message, current, total))
}

func (p *mockProgress) SetMessage(message string) {
	p.ui.mu.Lock()
	defer p.ui.mu.Unlock()
	
	p.message = message
	p.ui.messages = append(p.ui.messages, fmt.Sprintf("PROGRESS_MSG: %s", message))
}

func (p *mockProgress) Finish() {
	p.ui.mu.Lock()
	defer p.ui.mu.Unlock()
	
	p.ui.messages = append(p.ui.messages, fmt.Sprintf("PROGRESS_FINISH: %s", p.message))
}

func (p *mockProgress) Error(err error) {
	p.ui.mu.Lock()
	defer p.ui.mu.Unlock()
	
	p.ui.messages = append(p.ui.messages, fmt.Sprintf("PROGRESS_ERROR: %s - %v", p.message, err))
}
// SetConfirmResponse sets a default confirmation response
func (m *MockUIProvider) SetConfirmResponse(confirm bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Set default response for any confirm message
	m.confirms[""] = confirm
}
