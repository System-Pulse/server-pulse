package model

import "time"

// HealthMetrics holds the metrics for the system health check.
type HealthMetrics struct {
	IOWait          float64
	ContextSwitches uint64
	Interrupts      uint64
	StealTime       float64
	MajorFaults     uint64
	MinorFaults     uint64
}

// HealthScore represents the overall health score and any identified issues.
type HealthScore struct {
	Score           int
	Issues          []string
	Recommendations []string
	ChecksPerformed []string
}

type PerformanceTab int

const (
	SystemHealth PerformanceTab = iota
	InputOutput
	CPU
	Memory
	QuickTests
)

func (pt PerformanceTab) String() string {
	return []string{"System Health", "I/O", "CPU", "Memory", "Quick Tests"}[pt]
}

type PerformanceModel struct {
	SelectedItem           PerformanceTab
	Nav                    []string
	SubTabNavigationActive bool
	HealthMetrics          *HealthMetrics
	HealthScore            *HealthScore
	HealthLoading          bool
	IOMetrics              *IOMetrics
	IOLoading              bool
}

// DiskIOInfo represents I/O statistics for a single disk
type DiskIOInfo struct {
	Device      string
	ReadIOPS    uint64
	WriteIOPS   uint64
	ReadBytes   uint64
	WriteBytes  uint64
	ReadTime    uint64
	WriteTime   uint64
	QueueDepth  uint64
	Utilization float64
}

// ProcessIOInfo represents I/O statistics for a process
type ProcessIOInfo struct {
	PID        int32
	Command    string
	ReadIOPS   uint64
	WriteIOPS  uint64
	ReadBytes  uint64
	WriteBytes uint64
}

// IOMetrics holds comprehensive I/O performance metrics
type IOMetrics struct {
	Disks           []DiskIOInfo
	TopProcesses    []ProcessIOInfo
	TotalReadIOPS   uint64
	TotalWriteIOPS  uint64
	TotalReadBytes  uint64
	TotalWriteBytes uint64
	AverageLatency  float64
	LastUpdate      time.Time
}
