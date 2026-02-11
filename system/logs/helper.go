package logs

import (
	"strings"
	"time"
)

func parsePriority(priority string) int {
	switch priority {
	case "0":
		return 0 // Emergency
	case "1":
		return 1 // Alert
	case "2":
		return 2 // Critical
	case "3":
		return 3 // Error
	case "4":
		return 4 // Warning
	case "5":
		return 5 // Notice
	case "6":
		return 6 // Info
	case "7":
		return 7 // Debug
	default:
		return 6 // Default to Info
	}
}

func priorityToLevel(priority int) string {
	switch priority {
	case 0:
		return "EMERG"
	case 1:
		return "ALERT"
	case 2:
		return "CRIT"
	case 3:
		return "ERROR"
	case 4:
		return "WARN"
	case 5:
		return "NOTICE"
	case 6:
		return "INFO"
	case 7:
		return "DEBUG"
	default:
		return "INFO"
	}
}

func parseLogLine(line string) LogEntry {
	entry := LogEntry{
		FullText: line,
		Message:  line,
		Level:    "INFO",
	}

	// Try to extract timestamp (common formats)
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		// Try to parse timestamp
		timeStr := strings.Join(parts[:3], " ")
		if ts, err := time.Parse("Jan 2 15:04:05", timeStr); err == nil {
			entry.Timestamp = ts.AddDate(time.Now().Year(), 0, 0)
			entry.Message = strings.Join(parts[3:], " ")
		}
	}

	// Extract service/process name if present
	if len(parts) >= 4 {
		entry.Service = strings.TrimRight(parts[3], ":")
	}

	// Detect level from message content
	lowerMsg := strings.ToLower(line)
	if strings.Contains(lowerMsg, "error") || strings.Contains(lowerMsg, "fail") {
		entry.Level = "ERROR"
		entry.Priority = 3
	} else if strings.Contains(lowerMsg, "warn") {
		entry.Level = "WARN"
		entry.Priority = 4
	} else if strings.Contains(lowerMsg, "crit") {
		entry.Level = "CRIT"
		entry.Priority = 2
	}

	return entry
}

// convertTimeRange converts shorthand time ranges to journalctl format
// If the input is not a shorthand, pass it through as-is for journalctl
func convertTimeRange(timeRange string) string {
	switch timeRange {
	case "5m":
		return "5 minutes ago"
	case "1h":
		return "1 hour ago"
	case "24h":
		return "1 day ago"
	case "7d":
		return "7 days ago"
	case "30d":
		return "30 days ago"
	default:
		// Pass through custom values as-is (e.g., "1 minute ago", "2025-01-08")
		return timeRange
	}
}
