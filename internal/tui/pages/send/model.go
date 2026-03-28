package send

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/GrapeInTheTree/chiliz-cli/internal/domain"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/config"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/ethereum"
	"github.com/GrapeInTheTree/chiliz-cli/internal/tui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateChoosingWallet state = iota
	stateChoosingChain
	stateChoosingToken
	stateChoosingRecipient
	stateEnteringAmount
	stateConfirmSend
	stateSendingTx
	stateShowResult
)

type Model struct {
	state state

	// Data
	wallets  []domain.Wallet
	chains   []domain.Chain
	tokens   []domain.Token
	contacts []domain.Contact

	// Selections
	selectedWallet  *domain.Wallet
	selectedChain   *domain.Chain
	selectedToken   *domain.Token
	selectedContact *domain.Contact
	availableTokens []domain.Token

	// Input
	amountInput string

	// Result
	txHash string
	errMsg error

	cursor int
	width  int
	height int
}

func New(wallets []domain.Wallet, chains []domain.Chain, tokens []domain.Token, contacts []domain.Contact) Model {
	return Model{
		state:    stateChoosingWallet,
		wallets:  wallets,
		chains:   chains,
		tokens:   tokens,
		contacts: contacts,
		cursor:   0,
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
		case "ctrl+c":
			return m, tea.Quit
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
		default:
			if m.state == stateEnteringAmount {
				return m.handleAmountInput(msg.String())
			}
		}

	case txSentMsg:
		m.state = stateShowResult
		m.txHash = msg.hash
		m.errMsg = msg.err
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
		m.state = stateChoosingRecipient
		m.cursor = 0

	case stateChoosingRecipient:
		m.selectedContact = &m.contacts[m.cursor]
		m.state = stateEnteringAmount
		m.amountInput = ""
		m.cursor = 0

	case stateEnteringAmount:
		if m.amountInput != "" {
			m.state = stateConfirmSend
			m.cursor = 0
		}

	case stateConfirmSend:
		if m.cursor == 0 { // Confirm
			m.state = stateSendingTx
			return m, m.sendTxCmd()
		} else { // Cancel
			return m, func() tea.Msg { return BackMsg{} }
		}

	case stateShowResult:
		return m, func() tea.Msg { return BackMsg{} }
	}
	return m, nil
}

func (m Model) handleAmountInput(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "backspace":
		if len(m.amountInput) > 0 {
			m.amountInput = m.amountInput[:len(m.amountInput)-1]
		}
	default:
		if len(key) == 1 && (key >= "0" && key <= "9" || key == ".") {
			m.amountInput += key
		}
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
	case stateChoosingRecipient:
		return len(m.contacts) - 1
	case stateConfirmSend:
		return 1
	default:
		return 0
	}
}

func (m Model) sendTxCmd() tea.Cmd {
	return func() tea.Msg {
		privateKey, err := config.GetPrivateKey(m.selectedWallet.EnvKey)
		if err != nil {
			return txSentMsg{err: err}
		}

		var amount *big.Int
		var hash string

		if m.selectedToken.IsNative() {
			amount, err = ethereum.ParseAmount(m.amountInput)
			if err != nil {
				return txSentMsg{err: fmt.Errorf("invalid amount: %w", err)}
			}

			hash, err = ethereum.SendTransaction(
				privateKey,
				m.selectedChain.RPCURL,
				m.selectedChain.GetChainIDBigInt(),
				m.selectedContact.Address,
				amount,
			)
		} else {
			amount, err = ethereum.ParseTokenAmount(m.amountInput, m.selectedToken.Decimals)
			if err != nil {
				return txSentMsg{err: fmt.Errorf("invalid amount: %w", err)}
			}

			hash, err = ethereum.SendTokenTransaction(
				privateKey,
				m.selectedChain.RPCURL,
				m.selectedChain.GetChainIDBigInt(),
				m.selectedToken.Address,
				m.selectedContact.Address,
				amount,
			)
		}

		return txSentMsg{hash: hash, err: err}
	}
}

// Messages
type BackMsg struct{}

type txSentMsg struct {
	hash string
	err  error
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

	case stateChoosingRecipient:
		s.WriteString(style.Title.Render("Select Recipient") + "\n\n")
		for i, c := range m.contacts {
			if m.cursor == i {
				s.WriteString(style.Selected.Render("  👤 " + c.Name + "  "))
			} else {
				s.WriteString(style.Normal.Render("  👤 " + c.Name + "  "))
			}
			s.WriteString("\n")
		}

	case stateEnteringAmount:
		s.WriteString(style.Title.Render("Enter Amount ("+m.selectedToken.Symbol+")") + "\n\n")
		input := m.amountInput
		if input == "" {
			input = "0"
		}
		s.WriteString(style.Selected.Render(input + " █"))
		s.WriteString("\n\n" + style.Subtitle.Render("Press Enter to continue"))

	case stateConfirmSend:
		s.WriteString(style.Title.Render("Confirm Transaction") + "\n\n")
		s.WriteString(fmt.Sprintf("Send %s %s\n", m.amountInput, m.selectedToken.Symbol))
		s.WriteString(fmt.Sprintf("To: %s (%s)\n\n", m.selectedContact.Name, m.selectedContact.Address))

		options := []string{"✅ Confirm", "❌ Cancel"}
		for i, opt := range options {
			if m.cursor == i {
				s.WriteString(style.Selected.Render("  " + opt + "  "))
			} else {
				s.WriteString(style.Normal.Render("  " + opt + "  "))
			}
			s.WriteString("\n")
		}

	case stateSendingTx:
		s.WriteString(style.Title.Render("Sending Transaction...") + "\n\n")
		s.WriteString(style.Subtitle.Render("Please wait..."))

	case stateShowResult:
		if m.errMsg != nil {
			s.WriteString(style.Error.Render("Error: " + m.errMsg.Error()))
		} else {
			s.WriteString(style.Success.Render("Transaction Sent!"))
			s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Render(m.txHash))
		}
		s.WriteString("\n\nPress Enter to return")
	}

	return style.MenuContainer.Render(s.String())
}
