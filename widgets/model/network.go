package model

import (
	resource "github.com/System-Pulse/server-pulse/system/resource"
	"github.com/charmbracelet/bubbles/table"
)

type NetworkModel struct {
	// TODO: Implement NetworkModel
	NetworkTable    table.Model
	NetworkResource resource.NetworkInfo
	Nav             []string
	SelectedItem    ContainerTab
}
