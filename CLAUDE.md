# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`go-ethereum-butler` is a hybrid CLI+TUI application for managing multi-chain EVM transactions. Built with Cobra (CLI), Bubble Tea (TUI), and go-ethereum, it provides both scriptable CLI commands and an interactive keyboard-driven interface for balance checks, transaction queries, and token transfers across EVM-compatible chains.

Currently configured for Chiliz Chain with PEPPER token. New chains/tokens are added via JSON config files with zero code changes.

## Development Commands

```bash
# Run (TUI mode)
go run ./cmd/butler

# Run (CLI mode)
go run ./cmd/butler address 0x1234...
go run ./cmd/butler tx 0xabcd... --json
go run ./cmd/butler block latest
go run ./cmd/butler chain-info --json

# Build
go build -o butler ./cmd/butler

# Test
go test ./...

# Lint
go vet ./...
golangci-lint run

# Logs (TUI mode writes to butler.log)
tail -f butler.log

# Regenerate ERC-20 bindings after ABI changes
abigen --abi internal/infra/ethereum/abi/erc20.json \
       --pkg contracts --type ERC20 \
       --out internal/infra/ethereum/contracts/erc20.go
```

## CLI Commands

```
butler                              # Interactive TUI mode (no args)
butler address <addr>               # Address info: balance, nonce, tx history, tokens
butler tx <hash>                    # Transaction details with receipt
butler block [number|latest]        # Block information
butler chain-info                   # Chain status: latest block, gas price

Global flags:
  --chain <name>     Blockchain network (default: first in chains.json)
  --json             Machine-readable JSON output
  --config <path>    Config directory path
```

## Architecture

Three-layer clean architecture under `internal/`, with Cobra CLI routing on top:

```
cmd/butler/
  main.go                           Entry point: Cobra Execute()
  cmd/
    root.go                         Cobra root command + global flags + PersistentPreRunE
                                    Resolves config, chain, and explorer client
    tui.go                          TUI launcher (no-args fallback)
    address.go                      `butler address <addr>` — parallel RPC + Explorer fetch
    tx.go                           `butler tx <hash>` — tx + receipt lookup
    block.go                        `butler block [number]` — block by number
    chaininfo.go                    `butler chain-info` — chain status

internal/
  domain/
    models.go                       Pure domain structs (Chain, Token, Wallet, Contact)
                                    Chain.ExplorerAPIURL for per-chain explorer API
                                    Token.IsNative() detects native vs ERC-20
    output.go                       CLI output types: AddressInfo, TxDetail, BlockInfo,
                                    ChainStatus, TokenBalance, TxSummary
                                    All JSON-serializable with stable field names

  infra/
    config/config.go                Loads chains.json, tokens.json, contacts.json, .env
                                    ResolveConfigDir() cascade: --config > env > ~/.butler/ > CWD
                                    GetPrivateKey() reads key from env only at signing time

    ethereum/
      client.go                     RPC queries: GetBalance, GetNonce, GetCode, GetChainID,
                                    GetGasPrice, GetLatestBlockNumber, GetTransaction,
                                    GetTransactionReceipt, GetBlock, SendTransaction (EIP-1559),
                                    FormatBalance, ParseAmount, GetAddressFromPrivateKey
      erc20.go                      ERC-20: GetTokenBalance, SendTokenTransaction,
                                    FormatTokenBalance, ParseTokenAmount
                                    pow10() uses big.Int (safe for any decimal count)
      abi/erc20.json                Standard ERC-20 ABI
      contracts/erc20.go            Auto-generated Go bindings (abigen). Do not edit.

    explorer/
      etherscan.go                  Chiliscan/Etherscan-compatible API client (Routescan)
                                    GetTxList, GetTokenBalances, GetTokenTxList
                                    Rate-limited (2 req/sec), no API key needed
                                    Base URL per chain via Chain.ExplorerAPIURL

  output/
    formatter.go                    Print(jsonMode, v) — type switch for human/JSON output
                                    Human: formatted tables to stdout
                                    JSON: json.MarshalIndent to stdout

  tui/
    app.go                          Router model: manages currentPage, holds shared data
                                    (wallets, chains, tokens, contacts), routes messages
                                    to active page sub-model
    style/style.go                  Shared Lipgloss styles (Title, Selected, Error, Success, etc.)
    pages/
      mainmenu/model.go             Menu: Send Transaction, Check Balance, Exit
      balance/model.go              Balance check flow (4-state machine):
                                    Wallet -> Chain -> Token -> fetch balance -> result
      send/model.go                 Send flow (8-state machine):
                                    Wallet -> Chain -> Token -> Recipient -> Amount
                                    -> Confirm -> send tx -> result
```

### Key Design Patterns

**Hybrid CLI+TUI:** Cobra routes commands. No args = TUI (Bubbletea). Subcommands = CLI with stdout output. Single binary, same infra layer shared.

**Dual data sources:** RPC for real-time chain state (balance, nonce, blocks, tx by hash). Explorer API (Chiliscan/Routescan) for indexed data (tx history by address, token discovery). Explorer is optional — graceful degradation if unavailable.

**Nested Models (TUI):** Each TUI page is a self-contained Bubble Tea model with its own `Init`, `Update`, `View`. The router (`app.go`) delegates to the active page.

**Type-safe contract bindings:** ERC-20 interactions use `abigen`-generated Go code, not manual ABI encoding.

**Config-driven extensibility:** Chains, tokens, and contacts are pure JSON. Both TUI and CLI dynamically use these at startup.

**Async blockchain calls:** RPC operations run as concurrent goroutines (CLI uses sync.WaitGroup, TUI uses Bubble Tea commands).

### Data Source Strategy

| Data | Source | Notes |
|------|--------|-------|
| Native balance | RPC `eth_getBalance` | |
| ERC-20 balance | RPC `eth_call` (balanceOf) | Known tokens only |
| Nonce, code, gas price | RPC | |
| Tx by hash, receipt | RPC | |
| Block by number | RPC | |
| **Tx history by address** | **Explorer API** | RPC cannot do this |
| **All token holdings** | **Explorer API** | Token discovery requires indexer |

### Config Path Resolution

1. `--config /path/to/dir` flag
2. `BUTLER_CONFIG_DIR` environment variable
3. `~/.butler/` if `chains.json` exists there
4. Current working directory (default, backward compatible)

### Navigation & Input (TUI)

- `up`/`k`, `down`/`j`: cursor movement
- `enter`: confirm selection
- `esc`: back to main menu
- `ctrl+c`: quit
- Amount entry: `0-9`, `.`, `backspace`

## Extending the App

### No code changes needed
| What | File |
|------|------|
| Add EVM chain | `chains.json` (include `explorer_api_url` for tx history support) |
| Add ERC-20 token | `tokens.json` (decimals must match contract) |
| Add contact | `contacts.json` |

### Code changes needed
| What | Where |
|------|-------|
| Add CLI command | Create `cmd/butler/cmd/<name>.go`, register in `root.go` init() |
| Add wallet | `internal/infra/config/config.go` LoadWallets() + `.env` |
| Add TUI page | Create `internal/tui/pages/<name>/model.go`, register in `app.go` |
| Add blockchain query | Add to `internal/infra/ethereum/client.go` |
| Add explorer query | Add to `internal/infra/explorer/etherscan.go` |
| Add output type | Add struct to `internal/domain/output.go`, add case in `internal/output/formatter.go` |
| Add contract type | Place ABI in `abi/`, run `abigen`, use bindings in new `internal/infra/ethereum/<name>.go` |

## Key Dependencies

- `cobra` v1.10.2 - CLI framework
- `bubbletea` v1.3.10 - TUI framework
- `lipgloss` v1.1.0 - TUI styling
- `go-ethereum` v1.16.7 - Ethereum client, signing, ABI bindings
- `godotenv` v1.5.1 - .env loader

## Security

Private keys live only in `.env` (gitignored). They are loaded via `config.GetPrivateKey(envKey)` at the moment of transaction signing and never cached. The `.env.example` template shows the expected variable names. CLI read-only commands never access private keys.

## Token Handling Notes

- Native vs ERC-20 detection: `token.IsNative()` checks if address is empty/zero
- Native transfers use fixed 21000 gas; ERC-20 transfers use `EstimateGas()` + 20% buffer
- All gas pricing is EIP-1559 (`SuggestGasTipCap` + `SuggestGasPrice` with 1.2x buffer)
- Decimals are dynamic per token (18 for most, 6 for USDC/USDT, 8 for WBTC)
- `pow10()` uses `big.Int` — safe for any decimal count (no int64 overflow)

## Chiliscan API

- Base URL: `https://api.routescan.io/v2/network/mainnet/evm/88888/etherscan/api`
- Free tier: no API key, 2 req/sec, 10,000 calls/day
- Etherscan-compatible format (module/action query params)
- Set per chain via `explorer_api_url` in `chains.json`
