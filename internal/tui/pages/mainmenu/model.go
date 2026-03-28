package mainmenu

import (
	"strings"

	"github.com/GrapeInTheTree/chiliz-cli/internal/tui/style"
	tea "github.com/charmbracelet/bubbletea"
)

// Options
const (
	OptionSend    = "📤 Send Transaction"
	OptionBalance = "💰 Check Balance"
	OptionExit    = "🚪 Exit"
)

var options = []string{OptionSend, OptionBalance, OptionExit}

type Model struct {
	cursor int
	width  int
	height int
}

func New() Model {
	return Model{
		cursor: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(options)-1 {
				m.cursor++
			}
		case "enter":
			return m, func() tea.Msg {
				return SelectionMsg{Option: options[m.cursor]}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	s.WriteString(style.Title.Render("Main Menu") + "\n\n")

	for i, opt := range options {
		if m.cursor == i {
			s.WriteString(style.Selected.Render("  " + opt + "  "))
		} else {
			s.WriteString(style.Normal.Render("  " + opt + "  "))
		}
		s.WriteString("\n")
	}

	return style.MenuContainer.Render(s.String())
}

// Messages
type SelectionMsg struct {
	Option string
}
