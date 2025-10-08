package logs

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type LogSource int

const (
	LogSourceJournald LogSource = iota
	LogSourceSyslog
	LogSourceFile
)

type LogLevel string

const (
	LogLevelAll       LogLevel = "all"
	LogLevelEmergency LogLevel = "emerg"
	LogLevelAlert     LogLevel = "alert"
	LogLevelCritical  LogLevel = "crit"
	LogLevelError     LogLevel = "err"
	LogLevelWarning   LogLevel = "warning"
	LogLevelNotice    LogLevel = "notice"
	LogLevelInfo      LogLevel = "info"
	LogLevelDebug     LogLevel = "debug"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Service   string
	Message   string
	FullText  string
	Priority  int
}

type LogFilters struct {
	TimeRange   string // "1h", "24h", "7d", "custom"
	TimeStart   time.Time
	TimeEnd     time.Time
	Level       LogLevel
	Service     string
	SearchText  string
	Limit       int
}

type LogsInfos struct {
	Source      LogSource
	Entries     []LogEntry
	TotalCount  int
	HasMore     bool
	Filters     LogFilters
	ErrorMsg    string
}

type LogsMsg LogsInfos

type LogManager struct {
	CanUseSudo   bool
	SudoPassword string
}

func NewLogManager() *LogManager {
	return &LogManager{}
}

// DetectLogSystem detects which log system is available
func (lm *LogManager) DetectLogSystem() LogSource {
	// Check for systemd/journald
	cmd := exec.Command("systemctl", "--version")
	if err := cmd.Run(); err == nil {
		return LogSourceJournald
	}

	// Check for syslog files
	cmd = exec.Command("ls", "/var/log/syslog")
	if err := cmd.Run(); err == nil {
		return LogSourceSyslog
	}

	cmd = exec.Command("ls", "/var/log/messages")
	if err := cmd.Run(); err == nil {
		return LogSourceSyslog
	}

	return LogSourceJournald // Default to journald
}

// GetSystemLogs retrieves system logs based on filters
func (lm *LogManager) GetSystemLogs(filters LogFilters) tea.Cmd {
	return func() tea.Msg {
		source := lm.DetectLogSystem()

		var entries []LogEntry
		var err error

		switch source {
		case LogSourceJournald:
			entries, err = lm.getJournalLogs(filters)
		case LogSourceSyslog:
			entries, err = lm.getSyslogEntries(filters)
		default:
			entries, err = lm.getJournalLogs(filters)
		}

		if err != nil {
			return LogsMsg(LogsInfos{
				Source:     source,
				Entries:    []LogEntry{},
				TotalCount: 0,
				Filters:    filters,
				ErrorMsg:   fmt.Sprintf("Error loading logs: %v", err),
			})
		}

		return LogsMsg(LogsInfos{
			Source:     source,
			Entries:    entries,
			TotalCount: len(entries),
			HasMore:    len(entries) >= filters.Limit,
			Filters:    filters,
			ErrorMsg:   "",
		})
	}
}

// getJournalLogs retrieves logs from systemd journal
func (lm *LogManager) getJournalLogs(filters LogFilters) ([]LogEntry, error) {
	args := []string{}

	// Time range - convert shorthand to journalctl format
	if filters.TimeRange != "" && filters.TimeRange != "custom" {
		timeArg := convertTimeRange(filters.TimeRange)
		if timeArg != "" {
			args = append(args, "--since", timeArg)
		}
	} else if !filters.TimeStart.IsZero() {
		args = append(args, "--since", filters.TimeStart.Format("2006-01-02 15:04:05"))
		if !filters.TimeEnd.IsZero() {
			args = append(args, "--until", filters.TimeEnd.Format("2006-01-02 15:04:05"))
		}
	}

	// Priority/Level
	if filters.Level != LogLevelAll {
		args = append(args, "--priority", string(filters.Level))
	}

	// Service/Unit
	if filters.Service != "" {
		args = append(args, "--unit", filters.Service)
	}

	// Limit
	if filters.Limit > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", filters.Limit))
	} else {
		args = append(args, "-n", "100")
	}

	// Output format JSON for easier parsing
	args = append(args, "-o", "json", "--no-pager")

	var cmd *exec.Cmd
	useSudo := lm.CanUseSudo && lm.SudoPassword != ""
	if useSudo {
		sudoArgs := append([]string{"-S", "journalctl"}, args...)
		cmd = exec.Command("sudo", sudoArgs...)
		cmd.Stdin = strings.NewReader(lm.SudoPassword + "\n")
	} else {
		cmd = exec.Command("journalctl", args...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to simple format if JSON fails
		return lm.getJournalLogsSimple(filters)
	}

	return lm.parseJournalJSON(output, filters)
}

// parseJournalJSON parses journalctl JSON output
func (lm *LogManager) parseJournalJSON(output []byte, filters LogFilters) ([]LogEntry, error) {
	entries := []LogEntry{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var journalEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &journalEntry); err != nil {
			continue
		}

		entry := LogEntry{}

		// Timestamp (microseconds since epoch)
		if ts, ok := journalEntry["__REALTIME_TIMESTAMP"].(string); ok {
			if usec, err := time.ParseDuration(ts + "us"); err == nil {
				entry.Timestamp = time.Unix(0, 0).Add(usec)
			}
		}

		// Priority
		if priority, ok := journalEntry["PRIORITY"].(string); ok {
			entry.Priority = parsePriority(priority)
			entry.Level = priorityToLevel(entry.Priority)
		}

		// Service/Unit
		if unit, ok := journalEntry["_SYSTEMD_UNIT"].(string); ok {
			entry.Service = unit
		} else if unit, ok := journalEntry["SYSLOG_IDENTIFIER"].(string); ok {
			entry.Service = unit
		}

		// Message
		if msg, ok := journalEntry["MESSAGE"].(string); ok {
			entry.Message = msg
			entry.FullText = msg
		}

		// Apply search filter
		if filters.SearchText != "" {
			if !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(filters.SearchText)) {
				continue
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// getJournalLogsSimple fallback to simple text format
func (lm *LogManager) getJournalLogsSimple(filters LogFilters) ([]LogEntry, error) {
	args := []string{}

	if filters.TimeRange != "" && filters.TimeRange != "custom" {
		timeArg := convertTimeRange(filters.TimeRange)
		if timeArg != "" {
			args = append(args, "--since", timeArg)
		}
	}

	if filters.Level != LogLevelAll {
		args = append(args, "--priority", string(filters.Level))
	}

	if filters.Service != "" {
		args = append(args, "--unit", filters.Service)
	}

	if filters.Limit > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", filters.Limit))
	} else {
		args = append(args, "-n", "100")
	}

	args = append(args, "--no-pager")

	var cmd *exec.Cmd
	useSudo := lm.CanUseSudo && lm.SudoPassword != ""
	if useSudo {
		sudoArgs := append([]string{"-S", "journalctl"}, args...)
		cmd = exec.Command("sudo", sudoArgs...)
		cmd.Stdin = strings.NewReader(lm.SudoPassword + "\n")
	} else {
		cmd = exec.Command("journalctl", args...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("journalctl failed: %v, output: %s", err, string(output))
	}

	return lm.parseSimpleLogFormat(output, filters), nil
}

// parseSimpleLogFormat parses traditional log format
func (lm *LogManager) parseSimpleLogFormat(output []byte, filters LogFilters) []LogEntry {
	entries := []LogEntry{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Apply search filter
		if filters.SearchText != "" {
			if !strings.Contains(strings.ToLower(line), strings.ToLower(filters.SearchText)) {
				continue
			}
		}

		entry := parseLogLine(line)
		entries = append(entries, entry)
	}

	return entries
}

// getSyslogEntries retrieves logs from syslog files
func (lm *LogManager) getSyslogEntries(filters LogFilters) ([]LogEntry, error) {
	// Try /var/log/syslog first
	files := []string{"/var/log/syslog", "/var/log/messages"}

	for _, file := range files {
		args := []string{"tail", "-n"}
		if filters.Limit > 0 {
			args = append(args, fmt.Sprintf("%d", filters.Limit))
		} else {
			args = append(args, "100")
		}
		args = append(args, file)

		var cmd *exec.Cmd
		if lm.CanUseSudo && lm.SudoPassword != "" {
			sudoArgs := append([]string{"-S"}, args...)
			cmd = exec.Command("sudo", sudoArgs...)
			cmd.Stdin = strings.NewReader(lm.SudoPassword + "\n")
		} else {
			cmd = exec.Command(args[0], args[1:]...)
		}

		output, err := cmd.Output()
		if err == nil {
			return lm.parseSimpleLogFormat(output, filters), nil
		}
	}

	return nil, fmt.Errorf("no syslog files accessible")
}

// Helper functions

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
func convertTimeRange(timeRange string) string {
	switch timeRange {
	case "1h":
		return "1 hour ago"
	case "24h":
		return "1 day ago"
	case "7d":
		return "7 days ago"
	case "30d":
		return "30 days ago"
	default:
		return ""
	}
}
