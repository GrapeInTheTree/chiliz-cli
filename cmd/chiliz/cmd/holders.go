package cmd

import (
	"fmt"

	"github.com/GrapeInTheTree/chiliz-cli/internal/domain"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/config"
	"github.com/GrapeInTheTree/chiliz-cli/internal/output"
	"github.com/spf13/cobra"
)

var holdersCmd = &cobra.Command{
	Use:   "holders <token-address>",
	Short: "Top token holders with balances and total holder count",
	Long: `Query Chiliscan for top token holders of any ERC-20 token.

Examples:
  butler holders 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67
  butler holders 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := config.ResolveAddress(args[0], appCtx.Contacts)
		if err != nil {
			return err
		}

		if appCtx.Explorer == nil {
			return fmt.Errorf("holders info requires explorer API (set explorer_api_url in chains.json)")
		}

		holders, err := appCtx.Explorer.GetTokenHolders(addr, 1, 10)
		if err != nil {
			return err
		}

		count, _ := appCtx.Explorer.GetTokenHolderCount(addr)

		result := domain.HoldersResult{
			Token:      addr,
			TotalCount: count,
			Holders:    holders,
		}

		return output.Print(jsonOutput, result)
	},
}
