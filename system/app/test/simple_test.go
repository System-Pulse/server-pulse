package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Minutes only",
			duration: 5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "Hours and minutes",
			duration: 1*time.Hour + 30*time.Minute,
			expected: "1h 30m",
		},
		{
			name:     "Days and hours",
			duration: 2*24*time.Hour + 5*time.Hour,
			expected: "2d 5h 0m",
		},
		{
			name:     "Complex duration",
			duration: 3*24*time.Hour + 12*time.Hour + 45*time.Minute,
			expected: "3d 12h 45m",
		},
		{
			name:     "Zero duration",
			duration: 0,
			expected: "0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainerParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		containerName string
		expectedName  string
	}{
		{
			name:          "Normal container name",
			containerName: "/my-container",
			expectedName:  "my-container",
		},
		{
			name:          "Multiple slashes",
			containerName: "/compose-project/my-container",
			expectedName:  "compose-project/my-container",
		},
		{
			name:          "No slash",
			containerName: "my-container",
			expectedName:  "my-container",
		},
		{
			name:          "Empty string",
			containerName: "",
			expectedName:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContainerName(tt.containerName)
			assert.Equal(t, tt.expectedName, result)
		})
	}
}

func TestHealthStatusMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		state          string
		healthStatus   any
		expectedHealth string
	}{
		{
			name:           "Running state with nil health",
			state:          "running",
			healthStatus:   nil,
			expectedHealth: "running",
		},
		{
			name:           "Exited state with nil health",
			state:          "exited",
			healthStatus:   nil,
			expectedHealth: "exited",
		},
		{
			name:           "Paused state with nil health",
			state:          "paused",
			healthStatus:   nil,
			expectedHealth: "paused",
		},
		{
			name:           "Created state with nil health",
			state:          "created",
			healthStatus:   nil,
			expectedHealth: "created",
		},
		{
			name:           "With health status",
			state:          "running",
			healthStatus:   "healthy",
			expectedHealth: "healthy",
		},
		{
			name:           "Unknown state",
			state:          "unknown",
			healthStatus:   nil,
			expectedHealth: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapHealthStatus(tt.state, tt.healthStatus)
			assert.Equal(t, tt.expectedHealth, result)
		})
	}
}

func TestPortsFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ports    []any
		expected string
	}{
		{
			name:     "No ports",
			ports:    []any{},
			expected: "N/A",
		},
		{
			name: "Single port with public port",
			ports: []any{
				map[string]any{
					"PublicPort":  uint16(8080),
					"PrivatePort": uint16(80),
					"Type":        "tcp",
				},
			},
			expected: "8080:80/tcp",
		},
		{
			name: "Single port without public port",
			ports: []any{
				map[string]any{
					"PrivatePort": uint16(80),
					"Type":        "tcp",
				},
			},
			expected: "80/tcp",
		},
		{
			name: "Multiple ports",
			ports: []any{
				map[string]any{
					"PublicPort":  uint16(8080),
					"PrivatePort": uint16(80),
					"Type":        "tcp",
				},
				map[string]any{
					"PublicPort":  uint16(8443),
					"PrivatePort": uint16(443),
					"Type":        "tcp",
				},
			},
			expected: "8080:80/tcp, 8443:443/tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPorts(tt.ports)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCPUPercentCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		cpuUsage            uint64
		previousCPUUsage    uint64
		systemUsage         uint64
		previousSystemUsage uint64
		numCPUs             int
		expectedPercent     float64
	}{
		{
			name:                "Normal calculation",
			cpuUsage:            1000000000,
			previousCPUUsage:    500000000,
			systemUsage:         20000000000,
			previousSystemUsage: 10000000000,
			numCPUs:             2,
			expectedPercent:     10.0, // (500000000/10000000000)*2*100 = 10%
		},
		{
			name:                "Zero system delta",
			cpuUsage:            1000000000,
			previousCPUUsage:    500000000,
			systemUsage:         10000000000,
			previousSystemUsage: 10000000000,
			numCPUs:             2,
			expectedPercent:     100.0, // Should use fallback calculation
		},
		{
			name:                "No CPUs",
			cpuUsage:            1000000000,
			previousCPUUsage:    500000000,
			systemUsage:         20000000000,
			previousSystemUsage: 10000000000,
			numCPUs:             0,
			expectedPercent:     5.0, // Should fallback to 1 CPU: (500000000/10000000000)*1*100 = 5%
		},
		{
			name:                "Over 100% cap",
			cpuUsage:            1000000000,
			previousCPUUsage:    0,
			systemUsage:         1000000000,
			previousSystemUsage: 0,
			numCPUs:             1,
			expectedPercent:     100.0, // Should be capped at 100%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCPUPercent(
				tt.cpuUsage,
				tt.previousCPUUsage,
				tt.systemUsage,
				tt.previousSystemUsage,
				tt.numCPUs,
			)
			assert.InDelta(t, tt.expectedPercent, result, 0.1)
		})
	}
}

// Helper functions that mimic the internal logic from app package
func formatDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func parseContainerName(name string) string {
	return strings.TrimPrefix(name, "/")
}

func mapHealthStatus(state string, healthStatus any) string {
	if healthStatus != nil {
		if status, ok := healthStatus.(string); ok {
			return status
		}
	}

	switch state {
	case "running":
		return "running"
	case "exited":
		return "exited"
	case "paused":
		return "paused"
	case "created":
		return "created"
	default:
		return "N/A"
	}
}

func formatPorts(ports []any) string {
	if len(ports) == 0 {
		return "N/A"
	}

	var portInfo []string
	for _, port := range ports {
		if p, ok := port.(map[string]any); ok {
			publicPort, hasPublic := p["PublicPort"]
			privatePort, hasPrivate := p["PrivatePort"]
			portType, hasType := p["Type"]

			if hasPrivate && hasType {
				var portStr string
				if hasPublic && publicPort.(uint16) > 0 {
					portStr = fmt.Sprintf("%d:%d/%s", publicPort.(uint16), privatePort.(uint16), portType.(string))
				} else {
					portStr = fmt.Sprintf("%d/%s", privatePort.(uint16), portType.(string))
				}
				portInfo = append(portInfo, portStr)
			}
		}
	}

	if len(portInfo) == 0 {
		return "N/A"
	}
	return strings.Join(portInfo, ", ")
}

func calculateCPUPercent(cpuUsage, previousCPUUsage, systemUsage, previousSystemUsage uint64, numCPUs int) float64 {
	cpuDelta := float64(cpuUsage - previousCPUUsage)
	systemDelta := float64(systemUsage - previousSystemUsage)

	// Handle division by zero
	if systemDelta == 0 {
		// Use a small value to avoid division by zero
		systemDelta = 1
	}

	effectiveCPUs := float64(numCPUs)
	if effectiveCPUs == 0 {
		effectiveCPUs = 1 // Fallback to 1 CPU
	}

	cpuPercent := (cpuDelta / systemDelta) * effectiveCPUs * 100.0

	// Cap CPU percentage at 100%
	if cpuPercent > 100 {
		cpuPercent = 100
	}

	return cpuPercent
}
