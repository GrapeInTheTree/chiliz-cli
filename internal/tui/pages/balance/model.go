package balance

import (
	"math/big"
	"strings"

	"github.com/GrapeInTheTree/chiliz-cli/internal/domain"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/config"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/ethereum"
	"github.com/GrapeInTheTree/chiliz-cli/internal/tui/style"
	tea "github.com/charmbracelet/bubbletea"
)

// Internal states for the balance flow
type state int

const (
	stateChoosingWallet state = iota
	stateChoosingChain
	stateChoosingToken
	stateShowResult
)

type Model struct {
	state state

	// Data
	wallets []domain.Wallet
	chains  []domain.Chain
	tokens  []domain.Token

	// Selections
	selectedWallet  *domain.Wallet
	selectedChain   *domain.Chain
	selectedToken   *domain.Token
	availableTokens []domain.Token

	// Result
	balance       string
	walletAddress string
	errMsg        error

	cursor int
	width  int
	height int
}

func New(wallets []domain.Wallet, chains []domain.Chain, tokens []domain.Token) Model {
	return Model{
		state:   stateChoosingWallet,
		wallets: wallets,
		chains:  chains,
		tokens:  tokens,
		cursor:  0,
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
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.getMaxCursor() {
				m.cursor++
			}
		case "enter":
			return m.handleEnter()
		}

	case balanceRetrievedMsg:
		m.balance = msg.balance
		m.walletAddress = msg.address
		m.errMsg = msg.err
		m.state = stateShowResult
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateChoosingWallet:
		m.selectedWallet = &m.wallets[m.cursor]
		m.state = stateChoosingChain
		m.cursor = 0

	case stateChoosingChain:
		m.selectedChain = &m.chains[m.cursor]
		// Prepare tokens
		m.availableTokens = []domain.Token{
			{
				Symbol:   m.selectedChain.CurrencySymbol,
				Name:     "Native " + m.selectedChain.CurrencySymbol,
				Decimals: 18,
				ChainID:  m.selectedChain.ChainID,
			},
		}
		m.availableTokens = append(m.availableTokens, config.GetTokensForChain(m.tokens, m.selectedChain.ChainID)...)
		m.state = stateChoosingToken
		m.cursor = 0

	case stateChoosingToken:
		m.selectedToken = &m.availableTokens[m.cursor]
		return m, m.getBalanceCmd()

	case stateShowResult:
		return m, func() tea.Msg { return BackMsg{} }
	}
	return m, nil
}

func (m Model) getMaxCursor() int {
	switch m.state {
	case stateChoosingWallet:
		return len(m.wallets) - 1
	case stateChoosingChain:
		return len(m.chains) - 1
	case stateChoosingToken:
		return len(m.availableTokens) - 1
	default:
		return 0
	}
}

func (m Model) getBalanceCmd() tea.Cmd {
	return func() tea.Msg {
		// 1. Get Private Key
		privateKey, err := config.GetPrivateKey(m.selectedWallet.EnvKey)
		if err != nil {
			return balanceRetrievedMsg{err: err}
		}

		// 2. Derive Address
		address, err := ethereum.GetAddressFromPrivateKey(privateKey)
		if err != nil {
			return balanceRetrievedMsg{err: err}
		}

		var formatted string
		var balance *big.Int

		// 3. Fetch Balance
		if m.selectedToken.IsNative() {
			balance, err = ethereum.GetBalance(m.selectedChain.RPCURL, address)
			if err != nil {
				return balanceRetrievedMsg{err: err}
			}
			formatted = ethereum.FormatBalance(balance)
		} else {
			balance, err = ethereum.GetTokenBalance(
				m.selectedChain.RPCURL,
				m.selectedToken.Address,
				address,
				m.selectedToken.Decimals,
			)
			if err != nil {
				return balanceRetrievedMsg{err: err}
			}
			formatted = ethereum.FormatTokenBalance(balance, m.selectedToken.Decimals)
		}

		return balanceRetrievedMsg{balance: formatted, address: address}
	}
}

// Messages
type BackMsg struct{}

type balanceRetrievedMsg struct {
	balance string
	address string
	err     error
}

func (m Model) View() string {
	var s strings.Builder

	switch m.state {
	case stateChoosingWallet:
		s.WriteString(style.Title.Render("Select Wallet") + "\n\n")
		for i, w := range m.wallets {
			if m.cursor == i {
				s.WriteString(style.Selected.Render("  💼 " + w.Name + "  "))
			} else {
				s.WriteString(style.Normal.Render("  💼 " + w.Name + "  "))
			}
			s.WriteString("\n")
		}

	case stateChoosingChain:
		s.WriteString(style.Title.Render("Select Chain") + "\n\n")
		for i, c := range m.chains {
			if m.cursor == i {
				s.WriteString(style.Selected.Render("  ⛓️ " + c.Name + "  "))
			} else {
				s.WriteString(style.Normal.Render("  ⛓️ " + c.Name + "  "))
			}
			s.WriteString("\n")
		}

	case stateChoosingToken:
		s.WriteString(style.Title.Render("Select Token") + "\n\n")
		for i, t := range m.availableTokens {
			if m.cursor == i {
				s.WriteString(style.Selected.Render("  💎 " + t.Symbol + "  "))
			} else {
				s.WriteString(style.Normal.Render("  💎 " + t.Symbol + "  "))
			}
			s.WriteString("\n")
		}

	case stateShowResult:
		if m.errMsg != nil {
			s.WriteString(style.Error.Render("Error: " + m.errMsg.Error()))
		} else {
			s.WriteString(style.Success.Render("Balance: " + m.balance + " " + m.selectedToken.Symbol))
		}
		s.WriteString("\n\nPress Enter to return")
	}

	return style.MenuContainer.Render(s.String())
}
