# go-ethereum-butler

A Terminal User Interface (TUI) application for managing multi-chain EVM transactions.

## Features

- 🔗 Multi-chain support (currently: Chiliz Chain)
- 💸 Send native currency (ETH, CHZ, etc.) and ERC-20 token transactions
- 💰 Check wallet balances for both native and ERC-20 tokens
- 💎 Custom ERC-20 token support with configurable decimals
- 📖 Address book management
- 🔐 Secure private key handling (environment variables)
- ✨ Interactive keyboard-driven UI with token selection

## Quick Start

### 1. Setup Configuration

Create your `.env` file from the template:

```bash
cp .env.example .env
```

Edit `.env` and add your private keys (without `0x` prefix):

```ini
BUTLER_WALLET_MAIN=your_private_key_here
BUTLER_WALLET_TEST=your_test_private_key_here
```

### 2. Configure Chains, Tokens, and Contacts

**chains.json** - Add EVM-compatible chains:
```json
[
  {
    "name": "Chiliz Chain",
    "rpc_url": "https://rpc.ankr.com/chiliz",
    "chain_id": 88888,
    "currency_symbol": "CHZ",
    "logo_url": "https://example.com/chz-logo.png"
  }
]
```

**tokens.json** - Add ERC-20 tokens (optional):
```json
[
  {
    "symbol": "PEPPER",
    "name": "Pepper Token",
    "address": "0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67",
    "decimals": 18,
    "chain_id": 88888,
    "logo_url": ""
  }
]
```

**contacts.json** - Add your frequently used addresses:
```json
[
  {
    "name": "My Friend",
    "address": "0x..."
  }
]
```

### 3. Build and Run

```bash
# Build
go build -o butler ./cmd/butler

# Run
./butler
```

Or run directly:

```bash
go run ./cmd/butler
```

## Usage

### Navigation

- **Up/Down arrows** or **j/k** - Navigate menu items
- **Enter** - Select/Confirm
- **Esc** - Go back to main menu
- **q** or **Ctrl+C** - Quit

### Main Menu

1. **Send Transaction** - Interactive flow to send native currency or ERC-20 tokens
   - Select wallet → Select chain → Select token (Native/ERC-20) → Select recipient → Enter amount → Confirm
2. **Check Balance** - View wallet balance for any token on any chain
   - Select wallet → Select chain → Select token (Native/ERC-20) → View balance
3. **Exit** - Quit the application

## Architecture
 
The project follows **Standard Go Project Layout** and **Clean Architecture** principles.
 
```
go-ethereum-butler/
├── cmd/butler/          # Application entry point
├── internal/
│   ├── domain/          # Pure business logic (Models)
│   ├── infra/           # Infrastructure (Ethereum, Config)
│   │   ├── config/      # Configuration loading
│   │   └── ethereum/    # Blockchain client & contracts
│   │       ├── abi/     # Raw ABI JSON files
│   │       └── contracts/ # Generated Go bindings (abigen)
│   └── tui/             # Terminal User Interface
│       ├── app.go       # Main Router (Bubble Tea)
│       ├── style/       # Shared styles
│       └── pages/       # Independent page components
│           ├── mainmenu/
│           ├── balance/
│           └── send/
├── chains.json          # Chain configurations
├── tokens.json          # ERC-20 token configurations
├── contacts.json        # Address book
└── .env                 # Private keys (DO NOT COMMIT)
```
 
### Design Patterns
 
#### Clean Architecture
- **Domain Layer** (`internal/domain`): Contains pure data structures (`Chain`, `Token`, `Wallet`) with no external dependencies.
- **Infrastructure Layer** (`internal/infra`): Handles external concerns like file I/O and blockchain RPC calls.
- **Presentation Layer** (`internal/tui`): Handles UI rendering and user input, depending only on Domain and Infrastructure.
 
#### Nested Models (Bubble Tea)
Instead of a single monolithic model, the UI is broken down into smaller, composable models:
- **`app.go`**: The parent model that acts as a router. It switches between different page models based on user navigation.
- **Pages**: Each screen (Menu, Balance, Send) is a self-contained Bubble Tea model with its own `Init`, `Update`, and `View`.
 
This makes the codebase highly scalable. To add a new feature, you simply create a new package in `internal/tui/pages/` and register it in `app.go`.

See `CLAUDE.md` for detailed developer documentation.

## Security

- Private keys are **never** stored in code or configuration files
- Keys are only loaded from environment variables at signing time
- `.env` is automatically excluded from git via `.gitignore`
- Always keep your `.env` file secure and never share it

## Adding New Chains

Simply add a new entry to `chains.json`:

```json
{
  "name": "Your Chain Name",
  "rpc_url": "https://rpc.yourchain.com",
  "chain_id": 12345,
  "currency_symbol": "TOKEN",
  "logo_url": "https://example.com/token-logo.png"
}
```

The TUI will automatically detect and display the new chain. The `logo_url` is used for the native token icon (optional).

## Adding New Tokens

Add ERC-20 tokens to `tokens.json`:

```json
{
  "symbol": "USDC",
  "name": "USD Coin",
  "address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
  "decimals": 6,
  "chain_id": 1,
  "logo_url": ""
}
```

**Key points:**
- `decimals`: Token decimal places (18 for most tokens, 6 for USDC/USDT)
- `chain_id`: Must match a chain in `chains.json`
- `logo_url`: Optional URL for token icon/logo
- Same token on different chains needs separate entries
- Native tokens (ETH, CHZ, etc.) are automatically available and use the chain's `logo_url`

## Requirements

- Go 1.25.1 or higher
- Access to EVM-compatible RPC endpoints

## Development Setup
 
### Installing Abigen
 
To work with smart contracts, you need the `abigen` tool installed:
 
```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
```
 
### Generating Bindings
 
If you update the ABI files in `internal/infra/ethereum/abi/`, regenerate the Go bindings:
 
```bash
abigen --abi internal/infra/ethereum/abi/erc20.json --pkg contracts --type ERC20 --out internal/infra/ethereum/contracts/erc20.go
```
 
## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - TUI styling
- [go-ethereum](https://github.com/ethereum/go-ethereum) - Ethereum client
- [godotenv](https://github.com/joho/godotenv) - Environment variable management

## License

MIT
