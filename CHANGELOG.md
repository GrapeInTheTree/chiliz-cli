# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added
- **`butler call <contract> <sig> [args...]`** — generic read-only smart contract queries via `eth_call`
  - Cast-style signature format: `"functionName(inputTypes)(outputTypes)"`
  - Supports address, uint/int (all sizes), bool, string, bytes input types
  - Decodes return values including slices (`address[]`, `uint256[]`)
  - Raw hex fallback when output types are omitted
  - Reusable ABI helper module (`abi_helper.go`) for future commands
- **`butler validators`** — Chiliz-exclusive validator set and staking status
  - Queries Staking system contract (0x...1000) for `getValidators()` + `getValidatorStatus()`
  - Parallel queries for all 13 validators (concurrent goroutines)
  - Displays status, total delegated, commission rate, total rewards
- **`butler version`** — displays version and commit hash
  - GoReleaser ldflags injection at build time
  - Local builds show "dev (none)", releases show "v0.4.0 (abc1234)"
- **`butler rpc <method> [params]`** — raw JSON-RPC escape hatch for arbitrary RPC calls
- **`butler staking <address>`** — personal staking positions per validator via StakingPool (0x...7001)
  - Parallel `getStakedAmount` + `claimableRewards` queries for all 13 validators
  - Filters to only show validators with active stakes
- **`butler token <contract>`** — token metadata via Chiliscan `tokeninfo` API
  - Name, symbol, type, decimals, total supply, price, social links, verification status
- **4byte method decoding** in `butler tx` — auto-resolves method selectors to function names
  - OpenChain API + local cache of common ERC-20 selectors
  - e.g., `0x0efe6a8b` → `deposit(address,uint256,uint256)`
- **CI pipeline** — GitHub Actions: build + vet + test on every push/PR
- **MIT LICENSE** file
- **`CallContract()`** RPC function — `eth_call` wrapper
- **`DecodeRawOutputs()`** — returns raw Go types (not strings) for programmatic use
- **22 unit tests** for ABI helper functions (ParseCallSignature, ConvertArg, FormatValue, BuildCalldata)
- **Contact name resolution** — all address commands accept names from `contacts.json`
  - Case-insensitive partial match: `butler address danial` → resolves to 0xef33...
  - Actionable error messages: "address or contact not found" with hint
- **Makefile** — `make build`, `make test`, `make vet`, `make clean`, `make run`
- **CONTRIBUTING.md** — development setup, project structure, PR process
- **GitHub templates** — bug report, feature request, PR template
- **Improved error messages** — config loading errors include actionable hints with docs link

## [0.2.0] - 2026-03-27

This release transforms butler from a TUI-only app into a **hybrid CLI+TUI tool** with automated release infrastructure.

### Added

**CLI Framework (Cobra)**
- `butler address <addr>` — comprehensive address info: native balance, nonce, contract detection, ERC-20 token holdings, and last 10 transactions
- `butler tx <hash>` — full transaction details: status, block, from/to, value, gas used/limit, fee, method ID, log count
- `butler block [number|latest]` — block info: hash, parent, timestamp, miner, gas usage, base fee, transaction count
- `butler chain-info` — chain status: name, chain ID, RPC URL, latest block number, current gas price
- `--json` flag on all commands for machine-readable output (AI agent / script friendly)
- `--chain <name>` flag for multi-chain selection (default: first chain in `chains.json`)
- `--config <path>` flag for custom config directory location
- Running `butler` with no subcommand launches the existing TUI mode (zero breaking changes)

**Chiliscan Explorer API Client** (`internal/infra/explorer/etherscan.go`)
- Etherscan-compatible API integration via Routescan for Chiliz Chain
- `GetTxList()` — transaction history for an address (not possible via standard RPC)
- `GetTokenBalances()` — discover all ERC-20 token holdings for an address
- `GetTokenTxList()` — ERC-20 transfer history for an address
- Built-in rate limiting at 2 req/sec (Chiliscan free tier: no API key, 10,000 calls/day)
- Graceful degradation: if a chain has no `explorer_api_url`, explorer sections are simply omitted

**New RPC Query Functions** (`internal/infra/ethereum/client.go`)
- `GetNonce()` — confirmed transaction count for an address
- `GetCode()` — contract bytecode at an address (empty for EOA)
- `GetChainID()` — chain ID from the connected RPC node
- `GetGasPrice()` — current suggested gas price
- `GetLatestBlockNumber()` — latest block number
- `GetTransaction()` — transaction lookup by hash (includes pending detection)
- `GetTransactionReceipt()` — receipt with status, gas used, and event logs
- `GetBlock()` — full block by number (pass nil for latest)

**Output System** (`internal/output/formatter.go`, `internal/domain/output.go`)
- Dual-mode formatter: human-readable tables (default) or JSON (`--json`)
- Stable JSON output types: `AddressInfo`, `TxDetail`, `BlockInfo`, `ChainStatus`, `TokenBalance`, `TxSummary`
- Relative time display in human mode (e.g., "4d ago", "7h ago")
- Value direction indicators in address view (+received / -sent)

**Config Path Resolution** (`internal/infra/config/config.go`)
- 4-level cascade: `--config` flag > `BUTLER_CONFIG_DIR` env > `~/.butler/` > current working directory
- Backward compatible: existing users who run from project root see no change

**Chain Model Extension** (`internal/domain/models.go`)
- `ExplorerAPIURL` field in `Chain` struct for per-chain block explorer API

**Release Pipeline**
- GoReleaser configuration (`.goreleaser.yml`): cross-compiles for linux/darwin x amd64/arm64
- GitHub Actions workflow (`.github/workflows/release.yml`): auto-triggers on `v*` tag push
- Homebrew tap: `brew tap GrapeInTheTree/tap && brew install butler`
- Binaries available on [GitHub Releases](https://github.com/GrapeInTheTree/go-ethereum-butler/releases)

### Fixed
- **`pow10()` integer overflow** in `erc20.go` — changed from `int64` to `*big.Int`. The previous implementation would silently overflow for tokens with >18 decimals (int64 max is ~9.2x10^18). Now uses `big.Int.Exp()` which is safe for any decimal count.
- **Log file permissions** — tightened from `0666` to `0600` (owner read/write only)

### Changed
- `cmd/butler/main.go` refactored from 36-line direct TUI launch to 3-line Cobra `Execute()` call
- Config loading functions (`LoadChains`, `LoadTokens`, `LoadContacts`) now resolve file paths via `configPath()` instead of hardcoded relative paths
- `.env` loading attempts config directory first, then falls back to current working directory
- slog output silenced in CLI mode (TUI mode continues logging to `butler.log`)

## [0.1.0] - 2024-11-19

Initial release. TUI-only application.

### Added
- Interactive TUI with Bubble Tea framework (Elm Architecture)
- Three-page navigation: Main Menu > Check Balance / Send Transaction
- Native currency (CHZ) balance checks via `eth_getBalance` RPC
- Native currency transfers with EIP-1559 dynamic fee transactions
- ERC-20 token balance checks via `abigen`-generated contract bindings
- ERC-20 token transfers with auto gas estimation
- Multi-wallet support (Main Wallet, Test Wallet) via `.env` private keys
- Address book management via `contacts.json`
- Config-driven chain/token/contact management via JSON files
- Chiliz Chain (chain ID 88888) with PEPPER token pre-configured
- Structured JSON logging to `butler.log` via `slog`
- Lipgloss-styled UI with cursor navigation (j/k, up/down, enter, esc)
