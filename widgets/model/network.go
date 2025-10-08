package model

import (
	"github.com/System-Pulse/server-pulse/system/network"
	resource "github.com/System-Pulse/server-pulse/system/resource"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
)

type NetworkModel struct {
	NetworkTable      table.Model
	NetworkResource   resource.NetworkInfo
	ConnectionsTable  table.Model
	Connections       []network.ConnectionInfo
	RoutesTable       table.Model
	DNSTable          table.Model
	Routes            []network.RouteInfo
	DNS               []network.DNSInfo
	Nav               []string
	SelectedItem      ContainerTab
	PingInput         textinput.Model
	TracerouteInput   textinput.Model
	PingResults       []network.PingResult
	TracerouteResult  network.TracerouteResult
	ConnectivityMode  ConnectivityMode
	AuthState         AuthenticationState
	AuthMessage       string
	IsRoot            bool
	SudoAvailable     bool
	CanRunSudo        bool
	AuthTimer         int
	PingLoading       bool
	TracerouteLoading bool
}

type ConnectivityMode int

const (
	ConnectivityModeNone ConnectivityMode = iota
	ConnectivityModePing
	ConnectivityModeTraceroute
)
