package model

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
}
