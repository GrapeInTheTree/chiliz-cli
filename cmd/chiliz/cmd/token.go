package cmd

import (
	"fmt"

	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/config"
	"github.com/GrapeInTheTree/chiliz-cli/internal/output"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token <contract-address>",
	Short: "Token metadata: name, supply, price, holders, links",
	Long: `Query Chiliscan for token metadata: name, symbol, supply, holders, price.

Examples:
  butler token 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67
  butler token 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := config.ResolveAddress(args[0], appCtx.Contacts)
		if err != nil {
			return err
		}

		if appCtx.Explorer == nil {
			return fmt.Errorf("token info requires explorer API (set explorer_api_url in chains.json)")
		}

		info, err := appCtx.Explorer.GetTokenInfo(addr)
		if err != nil {
			return err
		}

		return output.Print(jsonOutput, *info)
	},
}
