package widgets

import (
	"strings"

	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/charmbracelet/bubbles/table"
)

func (m Model) renderTable(table table.Model, placeholder string) string {
	content := strings.Builder{}
	m.Ui.SearchInput.Placeholder = placeholder

	if m.Ui.SearchMode {
		searchBar := v.SearchBarStyle.
			Render(m.Ui.SearchInput.View())
		content.WriteString(searchBar)
		content.WriteString("\n")
	}

	content.WriteString(table.View())

	return v.CardTableStyle.Render(content.String())
}
