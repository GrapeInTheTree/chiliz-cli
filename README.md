<p align="center">
  <h1 align="center">go-ethereum-butler</h1>
  <p align="center">
    <strong>Your personal EVM blockchain assistant, from the terminal.</strong>
  </p>
  <p align="center">
    A hybrid CLI+TUI tool built with Go for querying and managing multi-chain EVM transactions — optimized for Chiliz Chain, extensible to any EVM network.
  </p>
</p>

<p align="center">
  <a href="https://github.com/GrapeInTheTree/go-ethereum-butler/releases"><img src="https://img.shields.io/github/v/release/GrapeInTheTree/go-ethereum-butler" alt="Release"></a>
  <a href="https://goreportcard.com/report/github.com/GrapeInTheTree/go-ethereum-butler"><img src="https://goreportcard.com/badge/github.com/GrapeInTheTree/go-ethereum-butler" alt="Go Report Card"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
  <a href="https://github.com/GrapeInTheTree/go-ethereum-butler/releases"><img src="https://img.shields.io/github/downloads/GrapeInTheTree/go-ethereum-butler/total" alt="Downloads"></a>
  <a href="https://github.com/GrapeInTheTree/go-ethereum-butler/actions/workflows/ci.yml"><img src="https://github.com/GrapeInTheTree/go-ethereum-butler/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

---

## Why go-ethereum-butler?

Most EVM CLI tools are either too low-level (raw RPC calls) or too generic (Foundry's `cast` has no concept of saved chains, contacts, or token configs). Butler fills the gap:

- **Stateful** — remembers your chains, tokens, and contacts across sessions via JSON config
- **Dual Data Sources** — combines RPC (real-time balance, blocks) with Explorer API (transaction history, token discovery) in a single command
- **AI Agent Friendly** — `--json` output on every command, designed for `butler address 0x... --json | jq`
- **Chiliz-First** — built for a chain with no dedicated CLI tooling (Chiliz is a go-ethereum fork)
- **Hybrid** — same binary runs as a scriptable CLI or an interactive TUI

Think of it as `cast` meets a personal wallet manager.

## Install

### Homebrew (macOS / Linux)

```bash
brew tap GrapeInTheTree/tap
brew install butler
```

### Go Install

```bash
go install github.com/GrapeInTheTree/go-ethereum-butler/cmd/butler@latest
```

### Download Binary

Pre-built binaries for macOS (Intel/Apple Silicon) and Linux (amd64/arm64) on the [Releases](https://github.com/GrapeInTheTree/go-ethereum-butler/releases) page.

### Build From Source

```bash
git clone https://github.com/GrapeInTheTree/go-ethereum-butler.git
cd go-ethereum-butler
go build -o butler ./cmd/butler
```

## How It Works

```
butler address 0xC3B2...D2c49
    │
    ├── RPC (go-ethereum ethclient) ──────────────────> Chiliz Node
    │     ├── eth_getBalance         → native balance     (rpc.ankr.com/chiliz)
    │     ├── eth_getTransactionCount → nonce
    │     └── eth_getCode            → EOA or contract
    │
    └── Explorer API (Chiliscan/Routescan) ───────────> Indexer
          ├── txlist                  → recent transactions
          └── addresstokenbalance    → all ERC-20 holdings

    ──> output.Print(jsonMode, AddressInfo)
          ├── human-readable table (default)
          └── JSON (--json flag)
```

| Data | Source | Why |
|------|--------|-----|
| Native balance, nonce, code | RPC `eth_*` | Standard EVM methods |
| Tx by hash + receipt | RPC | Direct lookup |
| Block by number | RPC | Direct lookup |
| Gas price, chain ID | RPC | Real-time state |
| **Tx history by address** | **Explorer API** | No RPC method exists (`eth_getTransactionsByAddress` doesn't exist) |
| **All token holdings** | **Explorer API** | Token discovery requires an indexer |

## CLI Usage

```
butler                              Interactive TUI mode (no subcommand)
butler address <addr>               Address info: balance, nonce, tx history, token holdings
butler tx <hash>                    Transaction details with receipt
butler block [number|latest]        Block information
butler chain-info                   Chain status: latest block, gas price
butler call <contract> <sig> [args] Read-only contract call (eth_call)
butler validators                   Chiliz validator set and staking status
butler staking <addr>               Personal staking positions and rewards
butler token <contract>             Token metadata, price, and social links
butler rpc <method> [params]        Raw JSON-RPC escape hatch
butler version                      Print version and build info

Global flags:
  --chain <name>     Blockchain network (default: first in chains.json)
  --json             Machine-readable JSON output
  --config <path>    Config directory path
  -h, --help         Help for any command
```

### Contact Name Resolution

All address commands support contact names from `contacts.json`:

```bash
butler address danial           # resolves to 0xef33...d93c
butler staking kayen            # resolves to 0x86d3...8897
butler address 0xC3B2...D2c49   # direct address still works
```

### butler address

Comprehensive address overview — 5 concurrent fetches (3 RPC + 2 Explorer) for fast response.

```bash
$ butler address 0xC3B2A6D869868916b1f5D46f9b7C62eD2f1D2c49

  Address:  0xC3B2A6D869868916b1f5D46f9b7C62eD2f1D2c49
  Chain:    Chiliz Chain (88888)
  Balance:  5045.818772 CHZ
  Nonce:    310
  Type:     EOA

  Token Holdings:
    PEPPER     12340.000000

  Recent Transactions (last 10):
    Hash            Method       Value              Time
    -----------------------------------------------------------------
    0x9f97...9981   Transfer     +3631.087415       4d ago
    0x12a8...b3e5   Transfer     +316.752300        7d ago
    0x1a79...a414   Transfer     -316.752300        7d ago
    ...
```

### butler tx

Full transaction details including receipt, gas breakdown, and event logs.

```bash
$ butler tx 0x9f978b07484bb439e790079afe192e0a562a93a26c9d893ea8001ddda88e9981

  Tx Hash:    0x9f978b...88e9981
  Status:     success
  Block:      32492204
  Time:       2026-03-23 00:23:15 UTC
  From:       0xa3DF8880d1D1BfC5Bea208AC3f1662420B2E2657
  To:         0xC3B2A6D869868916b1f5D46f9b7C62eD2f1D2c49
  Value:      3631.087415 CHZ
  Gas Price:  5001.00 Gwei
  Gas Used:   21000 / 21000
  Tx Fee:     0.105021 CHZ
  Nonce:      5343
  Logs:       0 events
```

### butler block

```bash
$ butler block latest

  Block:       #32617854
  Hash:        0x3f1317eb...75aeb5a
  Time:        2026-03-27 09:05:54 UTC
  Miner:       0xc10ae5Cd2C63E4065f81E241c93237E06E12d41b
  Gas Used:    70012 / 30000000
  Base Fee:    2500.00 Gwei
  Txs:         1 transactions
```

### butler call

Generic read-only contract calls — query any smart contract function.

```bash
# ERC-20 totalSupply
$ butler call 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67 "totalSupply()(uint256)"
  8760093706618541126999998364250000

# Token name and decimals
$ butler call 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67 "name()(string)"
  PEPPER

# Chiliz validator list (system contract)
$ butler call 0x0000000000000000000000000000000000001000 "getValidators()(address[])"
  [0x8d9B6aB3..., 0xBf968b8a..., 0x31dB8118..., ...]

# Check if address is a validator
$ butler call 0x0000000000000000000000000000000000001000 "isValidator(address)(bool)" 0x8d9B6aB3Fe8EbF16d9242e48feFB89360fa62820
  true

# JSON output
$ butler call 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67 "totalSupply()(uint256)" --json
{
  "contract": "0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67",
  "method": "totalSupply()",
  "values": ["8760093706618541126999998364250000"],
  "raw": "0x000000000000000000000000000000000001afe7ed5e433cdebfb05bd38def90"
}

# Raw hex (no output types specified)
$ butler call 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67 "totalSupply()"
  0x000000000000000000000000000000000001afe7ed5e433cdebfb05bd38def90
```

### butler validators

Chiliz-exclusive: query the on-chain Staking contract for validator status.

```bash
$ butler validators

  Chiliz Chain Validators (13 active)

  #    Address         Status     Delegated              Commission   Rewards
  --------------------------------------------------------------------------------
  1    0x8d9B...2820   Active     293,155,118 CHZ        1.0%         47,834 CHZ
  2    0xBf96...0B5f   Active     245,812,003 CHZ        5.0%         39,201 CHZ
  3    0x31dB...B074   Active     256,792,572 CHZ        10.0%        40,115 CHZ
  ...
```

### butler chain-info

```bash
$ butler chain-info --json
{
  "name": "Chiliz Chain",
  "chain_id": 88888,
  "rpc_url": "https://rpc.ankr.com/chiliz",
  "latest_block": 32617854,
  "gas_price": "2501.00 Gwei",
  "currency_symbol": "CHZ"
}
```

### JSON Output

Every command supports `--json` for piping and AI agent consumption:

```bash
butler address 0xC3B2... --json | jq .native_balance
butler chain-info --json | jq .latest_block
butler tx 0x9f97... --json | jq '{status, fee: .tx_fee}'
```

## TUI Usage

Run `butler` with no arguments for interactive mode.

```
┌─────────────────────────────────────┐
│  🔗  GO-ETHEREUM-BUTLER  🔗        │
│  Multi-Chain EVM Transaction Manager│
│                                     │
│  ┌─────────────────────────────┐    │
│  │  Main Menu                  │    │
│  │                             │    │
│  │  📤 Send Transaction        │    │
│  │  💰 Check Balance      ◄   │    │
│  │  🚪 Exit                   │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
```

| Key | Action |
|-----|--------|
| `up` / `k` | Move cursor up |
| `down` / `j` | Move cursor down |
| `enter` | Select / Confirm |
| `esc` | Back to main menu |
| `ctrl+c` | Quit |
| `0-9`, `.` | Amount input (send flow) |
| `backspace` | Delete last character |

**Check Balance** — Select wallet > chain > token (native or ERC-20) > view balance

**Send Transaction** — Select wallet > chain > token > recipient > enter amount > confirm > tx hash

## Configuration

Butler uses JSON config files. All are gitignored by default — create your own from the examples below.

### chains.json

```json
[
  {
    "name": "Chiliz Chain",
    "rpc_url": "https://rpc.ankr.com/chiliz",
    "chain_id": 88888,
    "currency_symbol": "CHZ",
    "logo_url": "",
    "explorer_api_url": "https://api.routescan.io/v2/network/mainnet/evm/88888/etherscan/api"
  }
]
```

| Field | Required | Description |
|-------|:--------:|-------------|
| `name` | Yes | Display name, also used with `--chain` flag |
| `rpc_url` | Yes | Any EVM-compatible JSON-RPC endpoint |
| `chain_id` | Yes | EIP-155 chain ID |
| `currency_symbol` | Yes | Native token symbol (CHZ, ETH, etc.) |
| `logo_url` | No | Optional logo URL |
| `explorer_api_url` | No | Etherscan-compatible API URL. Enables tx history and token discovery. Without it, CLI still shows RPC-based data. |

### tokens.json

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

- `decimals` must match the contract (18 for most, 6 for USDC/USDT, 8 for WBTC)
- `chain_id` must match a chain in `chains.json`
- Native tokens (CHZ, ETH) are automatically available — no entry needed
- Same token on different chains needs separate entries

### contacts.json

```json
[
  { "name": "My Friend", "address": "0x..." },
  { "name": "Team Wallet", "address": "0x..." }
]
```

Used by the TUI send flow for recipient selection.

### .env

```ini
BUTLER_WALLET_MAIN=your_private_key_here_without_0x_prefix
BUTLER_WALLET_TEST=your_test_private_key_here
```

Copy from `.env.example`. Private keys are **never** logged, cached, or committed to git. CLI read-only commands (`address`, `tx`, `block`, `chain-info`) do not access private keys at all.

<details>
<summary>Config Directory Resolution</summary>

Butler searches for config files in this order:

1. `--config /path/to/dir` flag (explicit)
2. `BUTLER_CONFIG_DIR` environment variable
3. `~/.butler/` directory (if `chains.json` exists there)
4. Current working directory (default)

This allows installing butler globally via `brew` and keeping configs in `~/.butler/`.

</details>

## Architecture

```
cmd/butler/
  main.go                       Entry point: Cobra Execute()
  cmd/
    root.go                     Root command + global flags + PersistentPreRunE
    tui.go                      TUI launcher (no-args fallback)
    address.go                  butler address — parallel RPC + Explorer
    tx.go                       butler tx — tx + receipt lookup
    block.go                    butler block — block by number
    chaininfo.go                butler chain-info — chain status
    call.go                     butler call — generic contract reads (eth_call)

internal/
  domain/
    models.go                   Chain, Token, Wallet, Contact structs
    output.go                   AddressInfo, TxDetail, BlockInfo, ChainStatus, CallResult

  infra/
    config/config.go            JSON/env config loading + path resolution
    ethereum/
      client.go                 RPC: balance, nonce, code, blocks, tx, gas, call
      abi_helper.go             Dynamic ABI encoding/decoding (no abigen needed)
      erc20.go                  ERC-20: balance, transfer, formatting
      abi/erc20.json            Standard ERC-20 ABI
      contracts/erc20.go        Auto-generated Go bindings (abigen)
    explorer/
      etherscan.go              Chiliscan API: tx history, token discovery

  output/
    formatter.go                Human + JSON dual formatter

  tui/
    app.go                      Bubble Tea router
    style/style.go              Lipgloss styles
    pages/                      mainmenu, balance, send
```

<details>
<summary>Design Decisions</summary>

**Cobra + Bubbletea hybrid** — Cobra routes subcommands. No args = Bubbletea TUI. Single binary, same infra layer shared. This follows the pattern used by `lazygit`, `k9s`, and other Go CLI/TUI hybrids.

**Package-level appContext** — CLI commands share resolved config/chain/explorer via a package-level struct in `root.go`. Appropriate for a CLI tool (single execution path, no concurrency at the command level). Not suitable for a server; would need DI then.

**Each RPC function creates its own ethclient** — `Dial → defer Close → call → return`. For a CLI that runs one command and exits, connection pool overhead isn't worth the lifecycle management complexity.

**Explorer graceful degradation** — If a chain has no `explorer_api_url` or the API is down, `butler address` still shows balance, nonce, and contract status from RPC. Explorer sections are simply omitted.

**Output types in `domain/`** — `AddressInfo`, `TxDetail`, etc. are stable JSON contracts. Placing them in `domain/` keeps them framework-agnostic and reusable across CLI, TUI, and future API layers.

**pow10 uses big.Int** — The original `int64`-based `pow10()` would silently overflow for tokens with >18 decimals. Fixed to use `big.Int.Exp()` which is safe for any decimal count.

**Dynamic ABI via `abi_helper.go`** — `butler call` uses `abi.ParseSelector` + `Arguments.Pack/Unpack` for runtime ABI encoding without JSON files. The same module will be reused by future commands (validators, staking, event logs). `ParseCallSignature` splits cast-style signatures via parenthesis depth counting because `abi.ParseSelector` rejects trailing characters.

</details>

## Releasing

Releases are automated via GoReleaser + GitHub Actions.

```bash
# 1. Commit your changes
git add . && git commit -m "feat: ..."

# 2. Tag and push
git tag v0.3.0
git push && git push --tags

# 3. GitHub Actions automatically:
#    - Cross-compiles linux/darwin × amd64/arm64
#    - Creates GitHub Release with changelog and binaries
#    - Updates Homebrew formula in GrapeInTheTree/homebrew-tap
```

Users upgrade with `brew upgrade butler` or `go install ...@latest`.

<details>
<summary>Release Infrastructure Details</summary>

| File | Purpose |
|------|---------|
| `.goreleaser.yml` | Build matrix (linux/darwin × amd64/arm64), archive format, Homebrew tap config |
| `.github/workflows/release.yml` | Triggered on `v*` tag push, runs `goreleaser release --clean` |

**Required GitHub secrets:**
- `HOMEBREW_TAP_TOKEN` — Fine-grained PAT with `repo` scope on `GrapeInTheTree/homebrew-tap`
- `GITHUB_TOKEN` — Automatic, used for creating GitHub Releases

</details>

## Shell Completion

Butler supports tab completion for bash, zsh, and fish:

```bash
# Bash
butler completion bash > /etc/bash_completion.d/butler

# Zsh (add to ~/.zshrc: fpath=(~/.zsh/completions $fpath))
butler completion zsh > ~/.zsh/completions/_butler

# Fish
butler completion fish > ~/.config/fish/completions/butler.fish
```

## Development

```bash
go build -o butler ./cmd/butler    # Build
go test ./...                       # Test (22 cases)
go vet ./...                        # Lint
tail -f butler.log                  # TUI logs
```

<details>
<summary>Regenerate ERC-20 Bindings</summary>

If you update `internal/infra/ethereum/abi/erc20.json`:

```bash
go install github.com/ethereum/go-ethereum/cmd/abigen@latest
abigen --abi internal/infra/ethereum/abi/erc20.json \
       --pkg contracts --type ERC20 \
       --out internal/infra/ethereum/contracts/erc20.go
```

</details>

<details>
<summary>Extending the App</summary>

**No code changes needed:**

| What | File |
|------|------|
| Add EVM chain | `chains.json` (include `explorer_api_url` for tx history) |
| Add ERC-20 token | `tokens.json` (decimals must match contract) |
| Add contact | `contacts.json` |

**Code changes needed:**

| What | Where |
|------|-------|
| Add CLI command | Create `cmd/butler/cmd/<name>.go`, register in `root.go` `init()` |
| Add wallet | `internal/infra/config/config.go` `LoadWallets()` + `.env` |
| Add TUI page | Create `internal/tui/pages/<name>/model.go`, register in `app.go` |
| Add blockchain query | Add to `internal/infra/ethereum/client.go` |
| Add explorer query | Add to `internal/infra/explorer/etherscan.go` |
| Add output type | Add struct to `internal/domain/output.go`, add case in `internal/output/formatter.go` |
| Add contract type | Place ABI in `abi/`, run `abigen`, use bindings in new file |

</details>

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25+ |
| CLI Framework | [spf13/cobra](https://github.com/spf13/cobra) v1.10.2 |
| TUI Framework | [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) v1.3.10 |
| TUI Styling | [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) v1.1.0 |
| EVM Client | [ethereum/go-ethereum](https://github.com/ethereum/go-ethereum) v1.16.7 |
| Env Loader | [joho/godotenv](https://github.com/joho/godotenv) v1.5.1 |
| Release | [GoReleaser](https://goreleaser.com/) + GitHub Actions |
| Explorer API | [Chiliscan](https://chiliscan.com/) (Routescan, Etherscan-compatible) |

## Security

- Private keys live only in `.env` (gitignored) and are loaded on-demand at signing time via `config.GetPrivateKey()`
- Keys are never cached in memory, logged, or written to any file
- CLI read-only commands never access private keys
- Git history audited: no secrets have ever been committed
- Log file permissions set to `0600` (owner read/write only)
- `.env`, `chains.json`, `tokens.json`, `contacts.json` are all gitignored

## License

[MIT](LICENSE)
