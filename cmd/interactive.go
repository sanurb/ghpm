package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/sanurb/ghpm/internal/ui"
)

func InteractiveCmd() error {
	zone.NewGlobal()

	const pageSize = 10

	p := tea.NewProgram(
		ui.NewTuiModel(pageSize),  // pass a user-defined page size
		tea.WithAltScreen(),       // alt screen
		tea.WithMouseCellMotion(), // mouse motion
	)

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
