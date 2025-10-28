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
	return []string{"System Health", "I/O", "CPU", "Memory"}[pt]
}

type CPUTab int

const (
	CPUTabStateBreakdown CPUTab = iota
	CPUTabPerCore
	CPUTabSystemActivity
)

func (ct CPUTab) String() string {
	return []string{"CPU State Breakdown", "Per-Core Performance", "System Activity Metrics"}[ct]
}

type MemoryTab int

const (
	MemoryTabOverview MemoryTab = iota
	MemoryTabUsageBreakdown
	MemoryTabSwapAnalysis
	MemoryTabSystemMemory
)

func (mt MemoryTab) String() string {
	return []string{"Overview", "Usage Breakdown", "Swap Analysis", "System Memory"}[mt]
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
	CPUMetrics             *CPUMetrics
	CPULoading             bool
	CPUSelectedTab         CPUTab
	CPUSubTabActive        bool
	MemoryMetrics          *MemoryMetrics
	MemoryLoading          bool
	MemorySelectedTab      MemoryTab
	MemorySubTabActive     bool
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

// CPUStateBreakdown represents detailed CPU state information
type CPUStateBreakdown struct {
	User      float64
	System    float64
	Idle      float64
	IOWait    float64
	IRQ       float64
	SoftIRQ   float64
	Steal     float64
	Nice      float64
	Guest     float64
	GuestNice float64
}

// CPUCoreInfo represents information for a single CPU core
type CPUCoreInfo struct {
	CoreID      int
	Usage       float64
	Frequency   float64
	Temperature float64 // If available
}

// CPUMetrics holds comprehensive CPU performance metrics
type CPUMetrics struct {
	OverallUsage    float64
	StateBreakdown  CPUStateBreakdown
	Cores           []CPUCoreInfo
	ContextSwitches uint64
	Interrupts      uint64
	LoadAverage     [3]float64
	ProcessCount    int
	ThreadCount     int
	LastUpdate      time.Time
}

// MemoryMetrics holds comprehensive memory performance metrics
type MemoryMetrics struct {
	// Overview
	Total       uint64
	Used        uint64
	Available   uint64
	Free        uint64
	UsedPercent float64

	// Usage Breakdown
	ApplicationMem uint64 // Used - Buffers - Cached
	Buffers        uint64
	Cached         uint64
	Shared         uint64

	// Swap Analysis
	SwapTotal       uint64
	SwapUsed        uint64
	SwapFree        uint64
	SwapUsedPercent float64
	SwapCached      uint64

	// System Memory
	Dirty      uint64
	WriteBack  uint64
	Slab       uint64
	PageTables uint64

	LastUpdate time.Time
}
