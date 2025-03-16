package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanurb/ghpm/internal/ui"
)

func InteractiveCmd() error {
	p := tea.NewProgram(ui.NewTuiModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
