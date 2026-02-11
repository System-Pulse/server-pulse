package logs

import "time"

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
	TimeRange  string // "1h", "24h", "7d", "custom"
	TimeStart  time.Time
	TimeEnd    time.Time
	Level      LogLevel
	Service    string
	SearchText string
	Limit      int
}

type LogsInfos struct {
	Source     LogSource
	Entries    []LogEntry
	TotalCount int
	HasMore    bool
	Filters    LogFilters
	ErrorMsg   string
}

type LogsMsg LogsInfos

type LogManager struct {
	CanUseSudo   bool
	SudoPassword string
}
