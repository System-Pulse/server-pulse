package model

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
}
