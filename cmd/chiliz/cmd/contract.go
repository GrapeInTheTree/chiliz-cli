package cmd

import (
	"fmt"

	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/config"
	"github.com/GrapeInTheTree/chiliz-cli/internal/output"
	"github.com/spf13/cobra"
)

var contractCmd = &cobra.Command{
	Use:   "contract <address>",
	Short: "Contract info: name, compiler, deployer, verification status",
	Long: `Query Chiliscan for smart contract metadata.

Examples:
  butler contract 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67
  butler contract 0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := config.ResolveAddress(args[0], appCtx.Contacts)
		if err != nil {
			return err
		}

		if appCtx.Explorer == nil {
			return fmt.Errorf("contract info requires explorer API (set explorer_api_url in chains.json)")
		}

		info, err := appCtx.Explorer.GetContractInfo(addr)
		if err != nil {
			return err
		}

		return output.Print(jsonOutput, *info)
	},
}
