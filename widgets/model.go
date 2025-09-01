package widgets

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"

	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	resource "github.com/System-Pulse/server-pulse/system/resource"
)

const (
	progressBarWidth = 40
)

type model struct {
	// Données
	system    info.SystemInfo
	cpu       resource.CPUInfo
	memory    resource.MemoryInfo
	disks     []resource.DiskInfo
	network   resource.NetworkInfo
	processes []proc.ProcessInfo
	err       error

	// État UI
	loading         bool
	selectedTab     int
	selectedMonitor int
	isMonitorActive bool        
	activeView      int             
	searchInput     textinput.Model 
	searchMode      bool            
	tabs            Menu
	width           int
	height          int
	minWidth        int
	minHeight       int
	ready           bool
	spinner         spinner.Model
	viewport        viewport.Model
	cpuProgress     progress.Model
	memProgress     progress.Model
	swapProgress    progress.Model
	processTable    table.Model
	diskProgress    map[string]progress.Model
	progressBars    map[string]progress.Model

	lastUpdate       time.Time
	enableAnimations bool
}

type Menu struct {
	DashBoard []string
	Monitor   []string
}
