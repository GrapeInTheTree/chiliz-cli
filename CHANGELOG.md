# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added
- **`chiliz call <contract> <sig> [args...]`** — generic read-only smart contract queries via `eth_call`
  - Cast-style signature format: `"functionName(inputTypes)(outputTypes)"`
  - Supports address, uint/int (all sizes), bool, string, bytes input types
  - Decodes return values including slices (`address[]`, `uint256[]`)
  - Raw hex fallback when output types are omitted
  - Reusable ABI helper module (`abi_helper.go`) for future commands
- **`chiliz validators`** — Chiliz-exclusive validator set and staking status
  - Queries Staking (0x...1000) + Governance (0x...7002) system contracts
  - **APY estimation** from on-chain rewards data (delegator APY after commission)
  - **Voting Power %** — exact match with Chiliz staking web (delegated / votingSupply)
  - Staked amounts in M (millions) for readability
  - Semaphore (max 4 concurrent) + retry for RPC rate limit resilience
- **`chiliz init`** — automatic config setup in `~/.chiliz/`
  - Chiliz Mainnet (88888) + Spicy Testnet (88882) pre-configured
  - PEPPER token, empty contacts, .env.example template
- **`chiliz version`** — displays version and commit hash
  - GoReleaser ldflags injection at build time
  - Local builds show "dev (none)", releases show "v0.4.0 (abc1234)"
- **`chiliz rpc <method> [params]`** — raw JSON-RPC escape hatch for arbitrary RPC calls
- **`chiliz staking <address>`** — personal staking positions per validator via StakingPool (0x...7001)
  - Parallel `getStakedAmount` + `claimableRewards` queries for all 13 validators
  - Filters to only show validators with active stakes
- **`chiliz token <contract>`** — token metadata via Chiliscan `tokeninfo` API
  - Name, symbol, type, decimals, total supply, price, social links, verification status
- **4byte method decoding** in `chiliz tx` — auto-resolves method selectors to function names
  - OpenChain API + local cache of common ERC-20 selectors
  - e.g., `0x0efe6a8b` → `deposit(address,uint256,uint256)`
- **CI pipeline** — GitHub Actions: build + vet + test on every push/PR
- **MIT LICENSE** file
- **`CallContract()`** RPC function — `eth_call` wrapper
- **`DecodeRawOutputs()`** — returns raw Go types (not strings) for programmatic use
- **22 unit tests** for ABI helper functions (ParseCallSignature, ConvertArg, FormatValue, BuildCalldata)
- **`chiliz contract <address>`** — contract source code, compiler, deployer, verification status via Chiliscan API
- **`chiliz holders <token>`** — top token holders with balances + total holder count via Chiliscan API
- **`chiliz logs`** — on-chain event log filtering via RPC `eth_getLogs`
  - `--address`, `--event`, `--blocks`, `--from-block`, `--to-block` flags
  - Event signature auto-hashed to topic0 via keccak256
- **`chiliz address` internal transactions** — internal (trace) txns via Chiliscan API
- **Contact name resolution** — all address commands accept names from `contacts.json`
  - Case-insensitive partial match: `chiliz address danial` → resolves to 0xef33...
  - Actionable error messages: "address or contact not found" with hint
- **Makefile** — `make build`, `make test`, `make vet`, `make clean`, `make run`
- **CONTRIBUTING.md** — development setup, project structure, PR process
- **GitHub templates** — bug report, feature request, PR template
- **`chiliz init`** — automatic config setup in `~/.chiliz/`
  - Chiliz Mainnet (88888) + Spicy Testnet (88882) pre-configured
  - PEPPER token, empty contacts, .env.example template
  - Safe to re-run (skips existing files)
  - `chiliz --chain "spicy" chain-info` for testnet
- **Contact name resolution** — all address commands accept names from `contacts.json`
- **Improved error messages** — config loading errors include actionable hints with docs link
- **Improved CLI help** — root `--help` shows Quick Start, each command has descriptive Short text

### Fixed
- **Validators "unknown" status** — semaphore (max 4 concurrent) + retry to avoid RPC rate limiting. All 13 validators now show consistently.

## [0.2.0] - 2026-03-27

This release transforms chiliz from a TUI-only app into a **hybrid CLI+TUI tool** with automated release infrastructure.

### Added

**CLI Framework (Cobra)**
- `chiliz address <addr>` — comprehensive address info: native balance, nonce, contract detection, ERC-20 token holdings, and last 10 transactions
- `chiliz tx <hash>` — full transaction details: status, block, from/to, value, gas used/limit, fee, method ID, log count
- `chiliz block [number|latest]` — block info: hash, parent, timestamp, miner, gas usage, base fee, transaction count
- `chiliz chain-info` — chain status: name, chain ID, RPC URL, latest block number, current gas price
- `--json` flag on all commands for machine-readable output (AI agent / script friendly)
- `--chain <name>` flag for multi-chain selection (default: first chain in `chains.json`)
- `--config <path>` flag for custom config directory location
- Running `chiliz` with no subcommand launches the existing TUI mode (zero breaking changes)

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
- 4-level cascade: `--config` flag > `CHILIZ_CONFIG_DIR` env > `~/.chiliz/` > current working directory
- Backward compatible: existing users who run from project root see no change

**Chain Model Extension** (`internal/domain/models.go`)
- `ExplorerAPIURL` field in `Chain` struct for per-chain block explorer API

**Release Pipeline**
- GoReleaser configuration (`.goreleaser.yml`): cross-compiles for linux/darwin x amd64/arm64
- GitHub Actions workflow (`.github/workflows/release.yml`): auto-triggers on `v*` tag push
- Homebrew tap: `brew tap GrapeInTheTree/tap && brew install chiliz`
- Binaries available on [GitHub Releases](https://github.com/GrapeInTheTree/chiliz-cli/releases)

### Fixed
- **`pow10()` integer overflow** in `erc20.go` — changed from `int64` to `*big.Int`. The previous implementation would silently overflow for tokens with >18 decimals (int64 max is ~9.2x10^18). Now uses `big.Int.Exp()` which is safe for any decimal count.
- **Log file permissions** — tightened from `0666` to `0600` (owner read/write only)

### Changed
- `cmd/chiliz/main.go` refactored from 36-line direct TUI launch to 3-line Cobra `Execute()` call
- Config loading functions (`LoadChains`, `LoadTokens`, `LoadContacts`) now resolve file paths via `configPath()` instead of hardcoded relative paths
- `.env` loading attempts config directory first, then falls back to current working directory
- slog output silenced in CLI mode (TUI mode continues logging to `chiliz.log`)

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
- Structured JSON logging to `chiliz.log` via `slog`
- Lipgloss-styled UI with cursor navigation (j/k, up/down, enter, esc)
