package model

import "time"

type ContainerTab int

const (
	ContainerTabGeneral ContainerTab = iota
	ContainerTabCPU
	ContainerTabMemory
	ContainerTabNetwork
	ContainerTabDisk
	ContainerTabEnv
	// "Interface", "Connectivity", "Configuration", "Protocol Analysis"
	// ================================ //
	NetworkTabInterface     ContainerTab = ContainerTabGeneral
	NetworkTabConnectivity               = ContainerTabCPU
	NetworkTabConfiguration              = ContainerTabMemory
	NetworkTabProtocol                   = ContainerTabNetwork
	// ================================ //

	// "Health Checks", "Performances", "Logs"
	DiagnosticSecurityChecks  ContainerTab = ContainerTabGeneral
	DiagnosticTabPerformances ContainerTab = ContainerTabCPU
	DiagnosticTabLogs         ContainerTab = ContainerTabMemory
)

type ContainerMenuItem struct {
	Key         string
	Label       string
	Description string
	Action      string
}

type DataHistory struct {
	Points    []DataPoint
	MaxPoints int
}

type ChartConfig struct {
	Title      string
	MaxValue   float64
	Height     int
	Width      int
	ShowLabels bool
}

type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

type ChartType int

type ContainerViewState int
type ShellExecRequest struct {
	ContainerID string
}

type ContainerMenuState int

type Menu struct {
	DashBoard []string
	Monitor   []string
}

type ClearOperationMsg struct{}
