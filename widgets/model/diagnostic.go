package model

import (
	"github.com/System-Pulse/server-pulse/system/security"
	"github.com/charmbracelet/bubbles/table"
)

type DiagnosticModel struct {
	DiagnosticTable table.Model
	Nav             []string
	SelectedItem    ContainerTab
	SecurityManager *security.SecurityManager
	SecurityTable   table.Model
	SecurityChecks  []security.SecurityCheck
	CertificateInfo *security.CertificateInfos
}
