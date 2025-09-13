package model

import "github.com/charmbracelet/bubbles/table"

type DiagnosticModel struct {
	// TODO: Implement DiagnosticModel
	DiagnosticTable table.Model
	Nav             []string
	SelectedItem    ContainerTab
}
