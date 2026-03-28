package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/config"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/explorer"
	"github.com/spf13/cobra"
)

// Global flags
var (
	chainFlag  string
	jsonOutput bool
	configFlag string
)

// Resolved shared state available to all subcommands
var appCtx struct {
	Chain    domain.Chain
	Chains   []domain.Chain
	Tokens   []domain.Token
	Contacts []domain.Contact
	Explorer *explorer.Client // nil if chain has no explorer URL
}

var rootCmd = &cobra.Command{
	Use:   "butler",
	Short: "Multi-chain EVM wallet manager",
	Long:  "A CLI and TUI tool for managing multi-chain EVM transactions. Run without subcommands for interactive TUI mode.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// In CLI mode, silence slog output (TUI sets up its own file logger)
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

		// Resolve config directory
		config.ResolveConfigDir(configFlag)

		// Load configs
		chains, err := config.LoadChains()
		if err != nil {
			return fmt.Errorf("failed to load chains: %w", err)
		}
		appCtx.Chains = chains

		tokens, err := config.LoadTokens()
		if err != nil {
			return fmt.Errorf("failed to load tokens: %w", err)
		}
		appCtx.Tokens = tokens

		contacts, _ := config.LoadContacts()
		appCtx.Contacts = contacts

		// Resolve --chain flag to a Chain struct
		chain, err := resolveChain(chains, chainFlag)
		if err != nil {
			return err
		}
		appCtx.Chain = chain

		// Create explorer client if chain has an API URL
		if chain.ExplorerAPIURL != "" {
			appCtx.Explorer = explorer.NewClient(chain.ExplorerAPIURL)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// No subcommand: launch TUI
		return runTUI()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&chainFlag, "chain", "", "blockchain network name (default: first in chains.json)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "path to config directory")

	rootCmd.AddCommand(addressCmd)
	rootCmd.AddCommand(txCmd)
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(chainInfoCmd)
	rootCmd.AddCommand(callCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(validatorsCmd)
	rootCmd.AddCommand(rpcCmd)
	rootCmd.AddCommand(stakingCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// resolveChain finds a chain by name (case-insensitive partial match), or returns the first chain if no flag is set
func resolveChain(chains []domain.Chain, name string) (domain.Chain, error) {
	if len(chains) == 0 {
		return domain.Chain{}, fmt.Errorf("no chains configured in chains.json")
	}

	if name == "" {
		return chains[0], nil
	}

	nameLower := strings.ToLower(name)
	for _, c := range chains {
		if strings.ToLower(c.Name) == nameLower || strings.Contains(strings.ToLower(c.Name), nameLower) {
			return c, nil
		}
	}
	return domain.Chain{}, fmt.Errorf("chain %q not found in chains.json", name)
}
