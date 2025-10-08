package vars

import (
	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/bubbles/spinner"
)

const (
	ProgressBarWidth = 40
)

const (
	ContainerMenuHidden model.ContainerMenuState = iota
	ContainerMenuVisible
)

const (
	ContainerViewNone model.ContainerViewState = iota
	ContainerViewSingle
	ContainerViewLogs
	ContainerViewConfirmation
)

var (
	dashboard = []string{"Monitor", "Diagnostic", "Network", "Reporting"}
	monitor   = []string{"System", "Process", "Containers"}
	Menu      = model.Menu{
		DashBoard: dashboard,
		Monitor:   monitor,
	}
	NetworkNav    = []string{"Interface", "Connectivity", "Configuration", "Protocol Analysis"}
	DiagnosticNav = []string{"Security Checks", "Performances", "Logs"}

	ContainerMenuItems = []model.ContainerMenuItem{
		{Key: "o", Label: "Open detailed view", Description: "View detailed container information", Action: "open_single"},
		{Key: "l", Label: "View logs", Description: "Show container logs", Action: "logs"},
		{Key: "r", Label: "Restart", Description: "Restart container", Action: "restart"},
		{Key: "d", Label: "Delete", Description: "Remove container", Action: "delete"},
		{Key: "x", Label: "Remove", Description: "Force remove container", Action: "remove"},
		{Key: "s", Label: "Stop/Start", Description: "Toggle container state", Action: "toggle_start"},
		{Key: "p", Label: "Pause/Resume", Description: "Toggle pause state", Action: "toggle_pause"},
		{Key: "e", Label: "Exec shell", Description: "Open interactive shell", Action: "exec"},
		// {Key: "t", Label: "Top/Stats", Description: "View real-time metrics", Action: "stats"},
		// {Key: "i", Label: "Inspect", Description: "Show container configuration", Action: "inspect"},
		// {Key: "c", Label: "Commit", Description: "Create image from container", Action: "commit"},
	}

	ContainerTabs = []string{"General", "CPU", "MEM", "NET", "ENV"} // disk remove

	Spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
)
