package system

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tc := range tests {
		result := FormatBytes(tc.input)
		if result != tc.expected {
			t.Errorf("FormatBytes(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestGetStats(t *testing.T) {
	stats, err := GetStats()
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}

	if stats.Hostname == "" {
		t.Error("hostname should not be empty")
	}
	if stats.OS == "" {
		t.Error("OS should not be empty")
	}
	if stats.CPUCores <= 0 {
		t.Errorf("expected positive CPU cores, got %d", stats.CPUCores)
	}
	// MemTotal should be > 0 on any real system
	if stats.MemTotal == 0 {
		t.Error("expected non-zero MemTotal")
	}
	if stats.MemPercent < 0 || stats.MemPercent > 100 {
		t.Errorf("MemPercent out of range: %f", stats.MemPercent)
	}
	if stats.CPUPercent < 0 || stats.CPUPercent > 100 {
		t.Errorf("CPUPercent out of range: %f", stats.CPUPercent)
	}
}
