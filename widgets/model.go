package widgets

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"

	"github.com/System-Pulse/server-pulse/system/app"
	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	resource "github.com/System-Pulse/server-pulse/system/resource"
)

const (
	progressBarWidth = 40
)

type ContainerMenuState int

const (
	ContainerMenuHidden ContainerMenuState = iota
	ContainerMenuVisible
)

type ContainerViewState int

const (
	ContainerViewNone ContainerViewState = iota
	ContainerViewSingle
)

type ContainerTab int

const (
	ContainerTabGeneral ContainerTab = iota
	ContainerTabCPU
	ContainerTabMemory
	ContainerTabNetwork
	ContainerTabDisk
	ContainerTabEnv
)

type ContainerMenuItem struct {
	Key         string
	Label       string
	Description string
	Action      string
}

type model struct {
	// Données
	system    info.SystemInfo
	cpu       resource.CPUInfo
	memory    resource.MemoryInfo
	disks     []resource.DiskInfo
	network   resource.NetworkInfo
	processes []proc.ProcessInfo
	app       *app.DockerManager
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
	container       table.Model
	diskProgress    map[string]progress.Model
	progressBars    map[string]progress.Model

	// État du menu contextuel des conteneurs
	containerMenuState ContainerMenuState
	selectedContainer  *app.Container
	containerMenuItems []ContainerMenuItem
	selectedMenuItem   int

	// État de la vue détaillée du conteneur
	containerViewState ContainerViewState
	containerTab       ContainerTab
	containerTabs      []string
	containerDetails   *app.ContainerDetails

	lastUpdate       time.Time
	enableAnimations bool
	// -------------------
	// Historique des données pour les graphiques
	cpuHistory       DataHistory
	memoryHistory    DataHistory
	networkRxHistory DataHistory
	networkTxHistory DataHistory
	diskReadHistory  DataHistory
	diskWriteHistory DataHistory

	// Dernière mise à jour des graphiques
	lastChartUpdate time.Time
	// -------------------
}

type Menu struct {
	DashBoard []string
	Monitor   []string
}
