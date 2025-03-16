package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// WelcomeBoxStyle is used to render the introduction/welcome message.
var WelcomeBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2)

// This color is used for important highlights in the TUI.
var highlightColor = lipgloss.Color("201")

// basePadding is a small global padding used around text blocks.
var basePadding = 1

// BaseStyle is the foundation style used for the entire UI background.
var BaseStyle = lipgloss.NewStyle().
	Margin(1, 2)

// TitleStyle is for big, bold titles like "Repositories" or "Select an option".
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("63"))

// ErrorStyle is for rendering error messages in a warm but noticeable way.
var ErrorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("160"))

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("69"))

// DoneMessageStyle is for a final message, post-cloning, for example.
var DoneMessageStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("229"))

var MenuCursorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("212"))

var MenuItemStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("250"))

var MenuSelectedItemStyle = MenuItemStyle.Copy().
	Foreground(lipgloss.Color("0")).
	Background(lipgloss.Color("212"))

// Download spinner style:
var DownloadSpinnerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("63"))

// Done checkmark
var CheckMarkStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("42"))

// Final done message style
var DoneStyle = lipgloss.NewStyle().
	Margin(1, 2)

// An example style for the currently-downloading repo name.
var CurrentRepoStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("211"))
