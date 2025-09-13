package app

import (
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerManager struct {
	Cli *client.Client
}

type Container struct {
	ID        string
	Name      string
	Image     string
	Project   string
	Command   string
	CreatedAt string
	Status    string
	State     string // (created, running, paused, etc.)
	Health    string
	Ports     []container.Port
	PortsStr  string
}

type PortInfo struct {
	PublicPort  uint16
	PrivatePort uint16
	Type        string
	HostIP      string
}

type ContainerStats struct {
	CPUPercent    float64
	MemoryUsage   uint64
	MemoryLimit   uint64
	MemoryPercent float64
	NetworkRx     uint64
	NetworkTx     uint64
	BlockRead     uint64
	BlockWrite    uint64
}

type ContainerDetails struct {
	Container
	Stats           ContainerStats
	Environment     []string
	IPAddress       string
	Gateway         string
	HealthCheck     string
	Uptime          string
	Ports           []PortInfo
	NetworkSettings any
	HostConfig      any
	State           any
}

type HealthInfo struct {
	Status        string
	FailingStreak int
	LastCheck     time.Time
	Output        string
}

type MountInfo struct {
	Source      string
	Destination string
	Type        string
	ReadOnly    bool
}

type ContainerMsg []Container
type ContainerDetailsMsg ContainerDetails

type ContainerLogsMsg struct {
	ContainerID string
	Logs        string
	Error       error
}

type ContainerOperationMsg struct {
	ContainerID string
	Operation   string // "restart", "start", "stop", "pause", "unpause", "delete"
	Success     bool
	Error       error
}

type ContainerStatusMsg struct {
	ContainerID string
	Status      string
	IsRunning   bool
	IsPaused    bool
}

type ContainerStatsChanMsg struct {
	ContainerID string
	StatsChan   chan ContainerStatsMsg
}

type ExecShellMsg struct {
	ContainerID string
}

type ContainerStatsMsg struct {
	ContainerID string
	CPUPercent  float64
	MemPercent  float64
	MemUsage    uint64
	MemLimit    uint64
	NetRX       uint64
	NetTX       uint64
	DiskUsage   uint64
}
