package model

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type UIModel struct {
	Loading         bool
	SelectedTab     int
	SelectedMonitor int
	IsMonitorActive bool
	IsNetworkActive bool
	ActiveView      int
	SearchInput     textinput.Model
	SearchMode      bool
	Tabs            Menu
	Width           int
	Height          int
	MinWidth        int
	MinHeight       int
	Ready           bool
	Spinner         spinner.Model
	Viewport        viewport.Model
}
