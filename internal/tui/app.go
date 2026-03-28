package tui

import (
	"strings"

	"github.com/GrapeInTheTree/chiliz-cli/internal/domain"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/config"
	"github.com/GrapeInTheTree/chiliz-cli/internal/tui/pages/balance"
	"github.com/GrapeInTheTree/chiliz-cli/internal/tui/pages/mainmenu"
	"github.com/GrapeInTheTree/chiliz-cli/internal/tui/pages/send"
	"github.com/GrapeInTheTree/chiliz-cli/internal/tui/style"
	tea "github.com/charmbracelet/bubbletea"
)

type page int

const (
	pageMenu page = iota
	pageBalance
	pageSend
)

type Model struct {
	currentPage page

	// Sub-models
	menuModel    mainmenu.Model
	balanceModel balance.Model
	sendModel    send.Model

	// Shared Data
	wallets  []domain.Wallet
	chains   []domain.Chain
	tokens   []domain.Token
	contacts []domain.Contact

	width  int
	height int
}

func NewModel() Model {
	return Model{
		currentPage: pageMenu,
		menuModel:   mainmenu.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		wallets, _ := config.LoadWallets()
		chains, _ := config.LoadChains()
		contacts, _ := config.LoadContacts()
		tokens, _ := config.LoadTokens()
		return configLoadedMsg{wallets, chains, contacts, tokens}
	}
}

type configLoadedMsg struct {
	wallets  []domain.Wallet
	chains   []domain.Chain
	contacts []domain.Contact
	tokens   []domain.Token
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate size to all sub-models
		var cmd1, cmd2, cmd3 tea.Cmd
		var model tea.Model

		model, cmd1 = m.menuModel.Update(msg)
		m.menuModel = model.(mainmenu.Model)

		model, cmd2 = m.balanceModel.Update(msg)
		m.balanceModel = model.(balance.Model)

		model, cmd3 = m.sendModel.Update(msg)
		m.sendModel = model.(send.Model)

		return m, tea.Batch(cmd1, cmd2, cmd3)

	case configLoadedMsg:
		m.wallets = msg.wallets
		m.chains = msg.chains
		m.contacts = msg.contacts
		m.tokens = msg.tokens

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Route messages to current page
	switch m.currentPage {
	case pageMenu:
		newModel, newCmd := m.menuModel.Update(msg)
		m.menuModel = newModel.(mainmenu.Model)
		cmd = newCmd

		// Check for navigation events
		if selection, ok := msg.(mainmenu.SelectionMsg); ok {
			switch selection.Option {
			case mainmenu.OptionSend:
				m.currentPage = pageSend
				m.sendModel = send.New(m.wallets, m.chains, m.tokens, m.contacts)
				// Initialize size
				var model tea.Model
				model, _ = m.sendModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				m.sendModel = model.(send.Model)
			case mainmenu.OptionBalance:
				m.currentPage = pageBalance
				m.balanceModel = balance.New(m.wallets, m.chains, m.tokens)
				var model tea.Model
				model, _ = m.balanceModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				m.balanceModel = model.(balance.Model)
			case mainmenu.OptionExit:
				return m, tea.Quit
			}
		}

	case pageBalance:
		newModel, newCmd := m.balanceModel.Update(msg)
		m.balanceModel = newModel.(balance.Model)
		cmd = newCmd

		if _, ok := msg.(balance.BackMsg); ok {
			m.currentPage = pageMenu
		}

	case pageSend:
		newModel, newCmd := m.sendModel.Update(msg)
		m.sendModel = newModel.(send.Model)
		cmd = newCmd

		if _, ok := msg.(send.BackMsg); ok {
			m.currentPage = pageMenu
		}
	}

	return m, cmd
}

func (m Model) View() string {
	var content string

	switch m.currentPage {
	case pageMenu:
		content = m.menuModel.View()
	case pageBalance:
		content = m.balanceModel.View()
	case pageSend:
		content = m.sendModel.View()
	}

	// Wrap with global layout (Title, Footer)
	var s strings.Builder

	title := style.BigTitle.Render("🔗  GO-ETHEREUM-BUTLER  🔗")
	subtitle := style.Subtitle.Render("Multi-Chain EVM Transaction Manager")

	s.WriteString(style.Center.Width(m.width).Render(title))
	s.WriteString("\n")
	s.WriteString(style.Center.Width(m.width).Render(subtitle))
	s.WriteString("\n")

	s.WriteString(style.Center.Width(m.width).Render(content))

	return s.String()
}
