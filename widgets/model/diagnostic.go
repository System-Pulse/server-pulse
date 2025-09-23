package model

import (
	"github.com/System-Pulse/server-pulse/system/security"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
)

type DiagnosticModel struct {
	DiagnosticTable table.Model
	Nav             []string
	SelectedItem    ContainerTab
	SecurityManager *security.SecurityManager
	SecurityTable   table.Model
	SecurityChecks  []security.SecurityCheck
	CertificateInfo *security.CertificateInfos
	SSHRootInfo     *security.SSHRootInfos
	DomainInput     textinput.Model
	DomainInputMode bool
	OpenedPortsInfo *security.OpenedPortsInfos
}
