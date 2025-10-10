package performance

import (
	"github.com/System-Pulse/server-pulse/widgets/vars"
)

func RenderSystemHealth() string {
	return vars.CardStyle.Render("System Health view is not yet implemented.")
}

func RenderInputOutput() string {
	return vars.CardStyle.Render("I/O view is not yet implemented.")
}

func RenderCPU() string {
	return vars.CardStyle.Render("CPU view is not yet implemented.")
}

func RenderMemory() string {
	return vars.CardStyle.Render("Memory view is not yet implemented.")
}

func RenderQuickTests() string {
	return vars.CardStyle.Render("Quick Tests view is not yet implemented.")
}
