package main

import "github.com/charmbracelet/lipgloss"

var (
	styleHeader    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Padding(0, 1)
	styleFooter    = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Padding(0, 1)
	styleTitle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	styleMuted     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleBody      = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	styleUser      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114"))
	styleAssistant = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	styleSystem    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	styleOK        = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	styleFocus     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	styleSelected  = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	styleCursor    = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	styleBorder    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
)
