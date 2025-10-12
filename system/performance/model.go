package performance

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
