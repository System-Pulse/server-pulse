package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getMaxContentWidth calculates the optimal content width based on terminal size
func (m Model) getMaxContentWidth() int {
	// Set minimum and maximum content widths
	minWidth := 40 // Augmenté de 80 à 40 pour les petits terminaux
	maxWidth := 120

	// Protection contre les largeurs négatives ou nulles
	if m.Ui.Width <= 0 {
		return minWidth
	}

	// For narrow terminals, use most of the width
	if m.Ui.Width < minWidth {
		return max(minWidth, 20) // Valeur minimale absolue
	}

	// For wide terminals, cap the content width for readability
	if m.Ui.Width > maxWidth {
		return maxWidth
	}

	// For medium terminals, use 90% of width
	calculatedWidth := int(float64(m.Ui.Width) * 0.9)
	return max(calculatedWidth, minWidth)
}

// centerLayout centers the entire layout horizontally within the terminal
func (m Model) centerLayout(content string) string {
	if m.Ui.Width <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	var centeredLines []string

	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < m.Ui.Width {
			leftPadding := (m.Ui.Width - lineWidth) / 2
			if leftPadding > 0 {
				centeredLine := strings.Repeat(" ", leftPadding) + line
				centeredLines = append(centeredLines, centeredLine)
			} else {
				centeredLines = append(centeredLines, line)
			}
		} else {
			centeredLines = append(centeredLines, line)
		}
	}

	return strings.Join(centeredLines, "\n")
}

// centerContent centers a content block within a given width
func (m Model) centerContent(content string, maxWidth int) string {
	if content == "" || maxWidth <= 0 {
		return content
	}

	style := lipgloss.NewStyle().
		Width(maxWidth).
		Align(lipgloss.Center)

	return style.Render(content)
}

// wrapWithMaxWidth wraps content with a maximum width constraint
func (m Model) wrapWithMaxWidth(content string, maxWidth int) string {
	if maxWidth <= 0 {
		return content
	}

	style := lipgloss.NewStyle().
		Width(maxWidth)

	return style.Render(content)
}

// centerBlock centers a block of content with padding
func (m Model) centerBlock(content string, maxWidth int) string {
	if content == "" || maxWidth <= 0 {
		return content
	}

	// Calculate padding for centering
	contentWidth := lipgloss.Width(content)
	if contentWidth >= maxWidth {
		return content
	}

	leftPadding := (maxWidth - contentWidth) / 2
	rightPadding := maxWidth - contentWidth - leftPadding

	style := lipgloss.NewStyle().
		PaddingLeft(leftPadding).
		PaddingRight(rightPadding)

	return style.Render(content)
}

// getResponsiveWidth returns a responsive width based on terminal size and content type
func (m Model) getResponsiveWidth(contentType string) int {
	maxWidth := m.getMaxContentWidth()

	switch contentType {
	case "table":
		// Tables need more space
		return min(maxWidth, m.Ui.Width-4)
	case "chart":
		// Charts can be wider
		return min(maxWidth-10, m.Ui.Width-8)
	case "card":
		// Cards should be more compact
		return min(maxWidth-20, 60)
	case "progress":
		// Progress bars scale with content
		return min(maxWidth/2, 40)
	default:
		return maxWidth
	}
}
