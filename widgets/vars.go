package widgets

var (
	asciiArt = `
 ▗▄▄▖▗▄▄▄▖▗▄▄▖ ▗▖  ▗▖▗▄▄▄▖▗▄▄▖     ▗▄▄▖ ▗▖ ▗▖▗▖    ▗▄▄▖▗▄▄▄▖
▐▌   ▐▌   ▐▌ ▐▌▐▌  ▐▌▐▌   ▐▌ ▐▌    ▐▌ ▐▌▐▌ ▐▌▐▌   ▐▌   ▐▌
 ▝▀▚▖▐▛▀▀▘▐▛▀▚▖▐▌  ▐▌▐▛▀▀▘▐▛▀▚▖    ▐▛▀▘ ▐▌ ▐▌▐▌    ▝▀▚▖▐▛▀▀▘
▗▄▄▞▘▐▙▄▄▖▐▌ ▐▌ ▝▚▞▘ ▐▙▄▄▖▐▌ ▐▌    ▐▌   ▝▚▄▞▘▐▙▄▄▖▗▄▄▞▘▐▙▄▄▖
		`
	dashboard = []string{"Monitor", "Diagnostic", "Network", "Reporting"}
	monitor   = []string{"System", "Process", "Application"}
	menu      = Menu{
		DashBoard: dashboard,
		Monitor:   monitor,
	}

	// Menu contextuel des conteneurs
	containerMenuItems = []ContainerMenuItem{
		{Key: "o", Label: "Open detailed view", Description: "View detailed container information", Action: "open_single"},
		{Key: "l", Label: "View logs", Description: "Show container logs", Action: "logs"},
		{Key: "r", Label: "Restart", Description: "Restart container", Action: "restart"},
		{Key: "d", Label: "Delete", Description: "Remove container", Action: "delete"},
		{Key: "x", Label: "Remove", Description: "Force remove container", Action: "remove"},
		{Key: "s", Label: "Stop/Start", Description: "Toggle container state", Action: "toggle_start"},
		{Key: "p", Label: "Pause/Resume", Description: "Toggle pause state", Action: "toggle_pause"},
		{Key: "e", Label: "Exec shell", Description: "Open interactive shell", Action: "exec"},
		{Key: "t", Label: "Top/Stats", Description: "View real-time metrics", Action: "stats"},
		{Key: "i", Label: "Inspect", Description: "Show container configuration", Action: "inspect"},
		// {Key: "c", Label: "Commit", Description: "Create image from container", Action: "commit"},
	}

	containerTabs = []string{"General", "CPU", "MEM", "NET", "DISK", "ENV"}
)
