package style

import "github.com/charmbracelet/lipgloss"

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	BigTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("235")).
			Padding(1, 4).
			MarginBottom(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170"))

	Subtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			MarginBottom(2)

	Selected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Background(lipgloss.Color("235")).
			Padding(0, 2)

	Normal = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Padding(0, 2)

	Error = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196"))

	Success = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46"))

	Center = lipgloss.NewStyle().
		Align(lipgloss.Center)

	MenuContainer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			MarginTop(1)
)
