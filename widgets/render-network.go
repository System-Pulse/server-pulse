package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/widgets/auth"
	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) interfaces() string {
	content := strings.Builder{}

	statusIcon := "🔴"
	statusText := "Disconnected"
	statusColor := v.ErrorColor

	if m.Network.NetworkResource.Connected {
		statusIcon = "🟢"
		statusText = "Connected"
		statusColor = v.SuccessColor
	}

	statusLine := fmt.Sprintf("%s %s",
		statusIcon,
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText))
	content.WriteString(statusLine)

	if len(m.Network.NetworkResource.Interfaces) > 0 {
		tableContent := m.renderTable(m.Network.NetworkTable, "No network interfaces")
		content.WriteString(tableContent)
	}

	return v.CardNetworkStyle.
		Margin(0, 0, 0, 0).
		Padding(0, 1).
		Render(content.String())
}

func (m Model) renderNetwork() string {
	currentView := ""

	switch m.Network.SelectedItem {
	case model.NetworkTabInterface:
		currentView = m.interfaces()
	case model.NetworkTabConnectivity:
		currentView = m.renderConnectivityAnalysis()
	case model.NetworkTabConfiguration:
		currentView = m.renderNetworkConfiguration()
	case model.NetworkTabProtocol:
		currentView = m.renderProtocolAnalysis()
	}
	return currentView
}

func (m Model) renderProtocolAnalysis() string {
	content := strings.Builder{}

	// Authentication section
	if m.Network.AuthState == model.AuthRequired || m.Network.AuthState == model.AuthInProgress {
		authMessage := auth.GetAuthMessage(int(m.Network.AuthState), m.Network.AuthMessage)
		authStyle := auth.GetAuthStyle(int(m.Network.AuthState))

		content.WriteString(authStyle.Render(authMessage))
		content.WriteString("\n\n")
		content.WriteString(m.Network.AuthMessage)
		if m.Network.AuthState == model.AuthRequired {
			content.WriteString(auth.AuthPromptStyle.Render("Enter Password:"))
			content.WriteString("\n")
			content.WriteString(m.Diagnostic.Password.View())
			content.WriteString("\n\n")
			content.WriteString(auth.AuthInfoStyle.Render(auth.AuthInstructions))
		} else {
			content.WriteString(auth.AuthInProgressStyle.Render("⏳ " + m.Network.AuthMessage))
		}
		content.WriteString("\n\n")
		return v.CardNetworkStyle.Render(content.String())
	}

	if m.Network.AuthState == model.AuthFailed {
		authMessage := auth.GetAuthMessage(int(m.Network.AuthState), m.Network.AuthMessage)
		authStyle := auth.GetAuthStyle(int(m.Network.AuthState))

		content.WriteString(authStyle.Render(authMessage))
		content.WriteString("\n\n")
		content.WriteString(m.Network.AuthMessage)
		content.WriteString("\n\n")
		content.WriteString(auth.AuthInfoStyle.Render(auth.AuthRetryMessage))
	}

	if m.Network.AuthState == model.AuthSuccess && m.Network.AuthTimer > 0 {
		authMessage := auth.GetAuthMessage(int(m.Network.AuthState), m.Network.AuthMessage)
		authStyle := auth.GetAuthStyle(int(m.Network.AuthState))

		content.WriteString(authStyle.Render(authMessage))
		content.WriteString("\n\n")
	}

	// Header with connection count
	connectionCount := len(m.Network.Connections)

	// Connection statistics
	if connectionCount > 0 {
		stats := m.getConnectionStats()
		statsText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Render(fmt.Sprintf("TCP: %d | UDP: %d | Listening: %d | Established: %d",
				stats.tcp, stats.udp, stats.listening, stats.established))
		content.WriteString(statsText + "\n")
	}

	// Check if user has access to detailed network connections
	hasAccess := m.canAccessNetworkConnections()

	if !hasAccess {
		accessInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("🔒 Limited information available without admin privileges")
		content.WriteString(accessInfo + "\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Render("Press 'a' to authenticate for detailed connection information"))
	}

	tableContent := m.renderTable(m.Network.ConnectionsTable, "No active connections")
	content.WriteString(tableContent)

	return v.CardNetworkStyle.
		Margin(0, 0, 0, 0).
		Padding(0, 1).
		Render(content.String())
}

func (m Model) renderConnectivityAnalysis() string {
	content := strings.Builder{}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		MarginBottom(1).
		Render("Network Connectivity Tools")
	content.WriteString(header + "\n")

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press 'p' to ping, 't' to traceroute, 's' for speed test, 'c' to clear results")
	content.WriteString(instructions + "\n\n")

	// Input modes - always visible
	switch m.Network.ConnectivityMode {
	case model.ConnectivityModePing:
		content.WriteString("🔍 Ping: " + m.Network.PingInput.View() + "\n\n")
	case model.ConnectivityModeTraceroute:
		content.WriteString("🛣️  Traceroute: " + m.Network.TracerouteInput.View() + "\n\n")
	case model.ConnectivityModeInstallPassword:
		content.WriteString("🔐 Sudo Password: " + m.Network.TracerouteInput.View() + "\n\n")
	}

	if m.Network.PingLoading {
		loadingText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render(fmt.Sprintf("%s Pinging...", m.Ui.Spinner.View()))
		content.WriteString(loadingText + "\n\n")
	}

	if m.Network.TracerouteLoading {
		loadingText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render(fmt.Sprintf("%s Running traceroute...", m.Ui.Spinner.View()))
		content.WriteString(loadingText + "\n\n")
	}

	if m.Network.SpeedTestLoading {
		loadingText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render(fmt.Sprintf("%s Running speed test...", m.Ui.Spinner.View()))
		content.WriteString(loadingText + "\n\n")
	}

	// Build results content for pagination
	resultsContent := strings.Builder{}

	// Ping results
	if len(m.Network.PingResults) > 0 {
		pingTitle := lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("39")).
			Render("Ping Results")
		resultsContent.WriteString(pingTitle + "\n")

		for _, result := range m.Network.PingResults {
			statusIcon := "❌"
			statusColor := v.ErrorColor
			if result.Success {
				statusIcon = "✅"
				statusColor = v.SuccessColor
			}

			statusLine := fmt.Sprintf("%s %s - ", statusIcon, result.Target)
			if result.Success {
				statusLine += fmt.Sprintf("Latency: %v, Packet Loss: %.1f%%",
					result.Latency, result.PacketLoss)
			} else {
				statusLine += fmt.Sprintf("Error: %s", result.Error)
			}

			resultsContent.WriteString(lipgloss.NewStyle().Foreground(statusColor).Render(statusLine) + "\n")
		}
		resultsContent.WriteString("\n")
	}

	// Traceroute results
	for _, tracerouteResult := range m.Network.TracerouteResults {
		if tracerouteResult.Target != "" {
			tracerouteTitle := lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.Color("208")).
				Render(fmt.Sprintf("Traceroute to %s", tracerouteResult.Target))
			resultsContent.WriteString(tracerouteTitle + "\n")

			if tracerouteResult.Error != "" {
				resultsContent.WriteString(lipgloss.NewStyle().Foreground(v.ErrorColor).Render("Error: "+tracerouteResult.Error) + "\n")
			} else if len(tracerouteResult.Hops) > 0 {
				for _, hop := range tracerouteResult.Hops {
					hopLine := fmt.Sprintf("%2d. ", hop.HopNumber)

					if hop.IP != "" {
						hopLine += hop.IP
						if hop.Hostname != "" {
							hopLine += fmt.Sprintf(" (%s)", hop.Hostname)
						}
					} else {
						hopLine += "*"
					}

					if hop.Latency1 > 0 {
						hopLine += fmt.Sprintf("  %v", hop.Latency1)
					}

					resultsContent.WriteString(hopLine + "\n")
				}
			} else {
				resultsContent.WriteString("No route found\n")
			}
			resultsContent.WriteString("\n")
		}
	}

	// Speed test results
	if len(m.Network.SpeedTestResults) > 0 {
		for _, result := range m.Network.SpeedTestResults {
			speedTestTitle := lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.Color("200")).
				Render("🚀 Speed Test Results")
			resultsContent.WriteString(speedTestTitle + "\n")

			resultsContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(fmt.Sprintf("Server: %s", result.Server)) + "\n")
			resultsContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(fmt.Sprintf("Test Duration: %v", result.TestDuration)) + "\n\n")

			if result.PingResult != nil {
				pingSection := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("85")).
					Render("📡 Latency (Ping):")
				resultsContent.WriteString(pingSection + "\n")
				resultsContent.WriteString(fmt.Sprintf("  Average: %.2f ms\n", float64(result.PingResult.Average.Microseconds())/1000))
				resultsContent.WriteString(fmt.Sprintf("  Min: %.2f ms, Max: %.2f ms\n",
					float64(result.PingResult.Min.Microseconds())/1000,
					float64(result.PingResult.Max.Microseconds())/1000))
				resultsContent.WriteString(fmt.Sprintf("  Samples: %d\n\n", result.PingResult.Samples))
			}

			downloadSection := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("85")).
				Render("⬇️  Download Speed:")
			resultsContent.WriteString(downloadSection + "\t")
			resultsContent.WriteString(fmt.Sprintf("  %.2f Mbps\n", result.DownloadMbps))

			uploadSection := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("85")).
				Render("⬆️  Upload Speed:")
			resultsContent.WriteString(uploadSection + "\t")
			resultsContent.WriteString(fmt.Sprintf("  %.2f Mbps\n", result.UploadMbps))

			// Summary box
			summaryBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1).
				MarginTop(1).
				Render(fmt.Sprintf("📊 Summary: %.2f ms latency | %.2f Mbps down | %.2f Mbps up",
					float64(result.PingResult.Average.Microseconds())/1000,
					result.DownloadMbps,
					result.UploadMbps))
			resultsContent.WriteString(summaryBox)
		}
	}

	// Apply pagination only to results
	resultsLines := strings.Split(resultsContent.String(), "\n")
	totalResultsLines := len(resultsLines)

	// Add pagination instructions if results exceed page limit
	if totalResultsLines > m.Network.ConnectivityPerPage {
		paginationInstructions := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("Use '↑' and '↓' to navigate results pages")
		content.WriteString(paginationInstructions + "\n\n")
	}

	startIdx := m.Network.ConnectivityPage * m.Network.ConnectivityPerPage
	endIdx := min(startIdx+m.Network.ConnectivityPerPage, totalResultsLines)
	totalPages := (totalResultsLines + m.Network.ConnectivityPerPage - 1) / m.Network.ConnectivityPerPage

	// Add paginated results to main content
	for i := startIdx; i < endIdx; i++ {
		content.WriteString(resultsLines[i] + "\n")
	}

	// Add pagination info if needed
	if totalPages > 1 {
		paginationInfo := fmt.Sprintf("Page %d/%d (Results %d-%d of %d)",
			m.Network.ConnectivityPage+1, totalPages, startIdx+1, endIdx, totalResultsLines)
		paginationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)
		content.WriteString("\n" + paginationStyle.Render(paginationInfo))
	}

	return v.CardNetworkStyle.
		Margin(0, 0, 0, 0).
		Padding(0, 1).
		Render(content.String())
}

func (m Model) getConnectionStats() connectionStats {
	var stats connectionStats

	for _, conn := range m.Network.Connections {
		switch conn.Proto {
		case "tcp", "tcp6":
			stats.tcp++
		case "udp", "udp6":
			stats.udp++
		}

		switch conn.State {
		case "LISTEN":
			stats.listening++
		case "ESTAB":
			stats.established++
		}
	}

	return stats
}

func (m Model) renderNetworkConfiguration() string {
	content := strings.Builder{}

	if m.Network.RoutesTable.Focused() {
		routesTitle := lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("208")).
			MarginBottom(1).
			Render(fmt.Sprintf("▶ Routing Table (%d routes)", len(m.Network.Routes)))
		content.WriteString(routesTitle + "\n")

		if len(m.Network.Routes) > 0 {
			content.WriteString(m.Network.RoutesTable.View())
		} else {
			content.WriteString("No routing table entries found")
		}
	} else {
		dnsTitle := lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1).
			Render(fmt.Sprintf("▶ DNS Servers (%d servers)", len(m.Network.DNS)))
		content.WriteString(dnsTitle + "\n")

		if len(m.Network.DNS) > 0 {
			content.WriteString(m.Network.DNSTable.View())
		} else {
			content.WriteString("No DNS servers configured")
		}
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(fmt.Sprintf("\nCurrently viewing: %s | Press SPACE to switch",
			func() string {
				if m.Network.RoutesTable.Focused() {
					return "Routing Table"
				}
				return "DNS Servers"
			}()))
	content.WriteString(footer)

	return v.CardNetworkStyle.
		Margin(0, 0, 0, 0).
		Padding(1, 2).
		Render(content.String())
}
