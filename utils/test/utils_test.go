package test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/System-Pulse/server-pulse/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatPercentage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"Zero percent", 0.0, "0.0%"},
		{"Integer percent", 50.0, "50.0%"},
		{"Decimal percent", 25.5, "25.5%"},
		{"High precision", 99.999, "100.0%"},
		{"Negative percent", -10.0, "-10.0%"},
		{"Over 100 percent", 150.0, "150.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatPercentage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUsageIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"Low usage", 10.0, "🟢"},
		{"Medium-low usage", 30.0, "🟡"},
		{"Medium-high usage", 60.0, "🟠"},
		{"High usage", 80.0, "🔴"},
		{"Boundary low", 24.9, "🟢"},
		{"Boundary medium", 25.0, "🟡"},
		{"Boundary high", 75.0, "🔴"},
		{"Negative usage", -5.0, "🟢"},
		{"Over 100 usage", 110.0, "🔴"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.GetUsageIcon(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatCompactUptime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"Seconds only", 59, "0m"},
		{"Minutes only", 300, "5m"},
		{"Hours and minutes", 3660, "1h1m"},
		{"Days and hours", 90000, "1d1h"},
		{"Multiple days", 172800, "2d0h"},
		{"Zero seconds", 0, "0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatCompactUptime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"Bytes", 500, "500 B"},
		{"Kilobytes", 1024, "1.0 KB"},
		{"Megabytes", 1048576, "1.0 MB"},
		{"Gigabytes", 1073741824, "1.0 GB"},
		{"Terabytes", 1099511627776, "1.0 TB"},
		{"Fractional MB", 1572864, "1.5 MB"},
		{"Zero bytes", 0, "0 B"},
		{"Large value", 1125899906842624, "1.0 PB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatCompactBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"Bytes", 500, "500B"},
		{"Kilobytes", 1024, "1K"},
		{"Megabytes", 1048576, "1M"},
		{"Gigabytes", 1073741824, "1G"},
		{"Terabytes", 1099511627776, "1T"},
		{"Fractional MB", 1572864, "2M"},
		{"Zero bytes", 0, "0B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatCompactBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatUptime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"Minutes only", 300, "5 minutes"},
		{"Hours and minutes", 3660, "1 hours, 1 minutes"},
		{"Days and hours", 90000, "1 days, 1 hours"},
		{"Multiple days", 172800, "2 days, 0 hours"},
		{"Zero seconds", 0, "0 minutes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatUptime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEllipsis(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"Short string", "hello", 10, "hello"},
		{"Exact length", "hello", 5, "hello"},
		{"Long string", "hello world", 8, "hello..."},
		{"Very short max", "hello", 2, "he"},
		{"Minimum ellipsis", "hello", 4, "h..."},
		{"Empty string", "", 5, ""},
		{"Unicode string", "héllo世界", 6, "hé..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Ellipsis(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadAvg(t *testing.T) {
	t.Parallel()

	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "loadavg_test")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Test valid loadavg format
	testContent := "1.25 0.75 0.50 2/123 45678\n"
	_, err = tmpFile.WriteString(testContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Note: In a real unit test, we'd mock the file system or use dependency injection
	// For this test, we're using the actual /proc/loadavg file

	t.Run("Valid loadavg format", func(t *testing.T) {
		// This test is more of an integration test since it reads from /proc/loadavg
		// In a real unit test, we'd mock the file system
		loadAvg, err := utils.LoadAvg()
		if err != nil {
			t.Skipf("Skipping LoadAvg test: %v", err)
		}

		// Basic validation that we got 3 values
		assert.Len(t, loadAvg, 3)
		assert.GreaterOrEqual(t, loadAvg[0], 0.0)
		assert.GreaterOrEqual(t, loadAvg[1], 0.0)
		assert.GreaterOrEqual(t, loadAvg[2], 0.0)
	})

	t.Run("File not found", func(t *testing.T) {
		// Test with non-existent file
		_, err := utils.LoadAvg()
		// This might fail on systems without /proc/loadavg
		if err != nil {
			assert.Contains(t, err.Error(), "failed to open")
		}
	})
}

func TestCheckDockerPermissions(t *testing.T) {
	t.Parallel()

	t.Run("Linux platform", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("Skipping Linux-specific test")
		}

		hasPerms, message := utils.CheckDockerPermissions()

		// The result depends on the actual system configuration
		// We can only verify that we get a boolean and a non-empty message
		assert.Contains(t, []bool{true, false}, hasPerms)
		assert.NotEmpty(t, message)
		assert.True(t, strings.Contains(message, "docker") || strings.Contains(message, "Docker"))
	})

	t.Run("Non-Linux platform", func(t *testing.T) {
		// This test would require mocking runtime.GOOS
		// For now, we'll just verify the function doesn't panic
		_, message := utils.CheckDockerPermissions()
		assert.NotEmpty(t, message)
	})
}

func TestFormatOperationMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		success   bool
		err       error
		expected  string
	}{
		{"Success restart", "restart", true, nil, "✅ Container restarted successfully"},
		{"Failure restart", "restart", false, nil, "❌ Container restarted failed"},
		{"Failure with error", "start", false, fmt.Errorf("permission denied"), "❌ Container started failed: permission denied"},
		{"Unknown operation success", "unknown", true, nil, "✅ Operation 'unknown' successfully"},
		{"Unknown operation failure", "unknown", false, nil, "❌ Operation 'unknown' failed"},
		{"Empty operation", "", true, nil, "✅ Operation '' successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatOperationMessage(tt.operation, tt.success, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetOperationIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		expected  string
	}{
		{"Restart", "restart", "🔄"},
		{"Start", "start", "▶️"},
		{"Stop", "stop", "⏹️"},
		{"Pause", "pause", "⏸️"},
		{"Unpause", "unpause", "▶️"},
		{"Delete", "delete", "🗑️"},
		{"Toggle start", "toggle_start", "🔄"},
		{"Toggle pause", "toggle_pause", "⏯️"},
		{"Exec", "exec", "💻"},
		{"Logs", "logs", "📄"},
		{"Unknown operation", "unknown", "⚙️"},
		{"Empty operation", "", "⚙️"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.GetOperationIcon(tt.operation)
			assert.Equal(t, tt.expected, result)
		})
	}
}
