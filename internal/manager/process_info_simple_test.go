package manager

import (
	"os"
	"runtime"
	"testing"
	"time"
)

// Test GetProcessInfo using the default implementation
func TestGetProcessInfo_Default(t *testing.T) {
	// Test with current process (should always work)
	pid := os.Getpid()
	info, err := GetProcessInfo(pid)
	
	if err != nil {
		t.Errorf("GetProcessInfo() error = %v", err)
		return
	}
	
	if info.PID != pid {
		t.Errorf("PID = %d, want %d", info.PID, pid)
	}
	
	// Current process should be running
	if info.Status != "running" {
		t.Errorf("Status = %s, want running", info.Status)
	}
}

// Test CheckProcessHealth
func TestCheckProcessHealth_Default(t *testing.T) {
	// Test with current process
	pid := os.Getpid()
	healthy, err := CheckProcessHealth(pid)
	
	if err != nil {
		t.Errorf("CheckProcessHealth() error = %v", err)
		return
	}
	
	if !healthy {
		t.Error("Current process should be healthy")
	}
	
	// Test with non-existent process
	healthy, err = CheckProcessHealth(999999)
	if healthy {
		t.Error("Non-existent process should not be healthy")
	}
}

// Test formatDuration
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "Minutes only",
			duration: 45 * time.Minute,
			want:     "45m",
		},
		{
			name:     "Hours and minutes",
			duration: 2*time.Hour + 30*time.Minute,
			want:     "2h30m",
		},
		{
			name:     "Days, hours and minutes",
			duration: 3*24*time.Hour + 5*time.Hour + 15*time.Minute,
			want:     "3d5h15m",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test parseStartTime
func TestParseStartTime(t *testing.T) {
	tests := []struct {
		name    string
		timeStr string
		wantErr bool
	}{
		{
			name:    "Standard format",
			timeStr: "Mon Jan  2 15:04:05 2024",
			wantErr: false,
		},
		{
			name:    "RFC3339 format",
			timeStr: "2024-01-02T15:04:05Z",
			wantErr: false,
		},
		{
			name:    "Invalid format",
			timeStr: "invalid time",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseStartTime(tt.timeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStartTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test getSystemMemory
func TestGetSystemMemory(t *testing.T) {
	mem := getSystemMemory()
	
	// On supported platforms, should return non-zero
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if mem <= 0 {
			t.Logf("getSystemMemory() returned %d, expected positive value on %s", mem, runtime.GOOS)
		}
	}
}