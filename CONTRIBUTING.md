# Contributing to go-ethereum-butler

Contributions are welcome! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/GrapeInTheTree/go-ethereum-butler.git
cd go-ethereum-butler
make build
make test
```

## Project Structure

```
cmd/butler/cmd/     CLI commands (one file per command)
internal/domain/    Data models and output types
internal/infra/     Infrastructure (RPC, config, explorer API)
internal/output/    Human/JSON output formatter
internal/tui/       Bubble Tea interactive UI
```

## Adding a New CLI Command

1. Create `cmd/butler/cmd/<name>.go`
2. Define a `var <name>Cmd = &cobra.Command{...}` with `RunE`
3. Register in `root.go` `init()` via `rootCmd.AddCommand(<name>Cmd)`
4. If needed, add output type in `internal/domain/output.go`
5. Add formatter case in `internal/output/formatter.go`

## Code Style

- Follow existing patterns (Dial/Close for RPC, type switch for output)
- Use `fmt.Errorf("context: %w", err)` for error wrapping
- Keep CLI commands thin — business logic in `internal/`
- Support `--json` output on all commands

## Testing

```bash
make test       # Run all tests
make vet        # Static analysis
go test -v ./internal/infra/ethereum/   # Verbose single package
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Make your changes with clear commit messages
4. Run `make test && make vet` before submitting
5. Open a pull request against `main`

## Commit Message Convention

```
feat: Add new feature
fix: Fix a bug
docs: Documentation changes
test: Add or update tests
ci: CI/CD changes
chore: Maintenance tasks
```
