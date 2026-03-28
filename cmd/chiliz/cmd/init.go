package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Default chain configurations
var defaultChains = []map[string]interface{}{
	{
		"name":             "Chiliz Chain",
		"rpc_url":          "https://rpc.ankr.com/chiliz",
		"chain_id":         88888,
		"currency_symbol":  "CHZ",
		"logo_url":         "",
		"explorer_api_url": "https://api.routescan.io/v2/network/mainnet/evm/88888/etherscan/api",
	},
	{
		"name":             "Chiliz Spicy Testnet",
		"rpc_url":          "https://spicy-rpc.chiliz.com",
		"chain_id":         88882,
		"currency_symbol":  "CHZ",
		"logo_url":         "",
		"explorer_api_url": "https://api.routescan.io/v2/network/testnet/evm/88882/etherscan/api",
	},
}

var defaultTokens = []map[string]interface{}{
	{
		"symbol":   "PEPPER",
		"name":     "Pepper Token",
		"address":  "0x60F397acBCfB8f4e3234C659A3E10867e6fA6b67",
		"decimals": 18,
		"chain_id": 88888,
		"logo_url": "",
	},
}

var defaultContacts = []map[string]interface{}{}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up config in ~/.chiliz/ (Chiliz Mainnet + Spicy Testnet)",
	Long: `Create default config files in ~/.chiliz/ directory.

Sets up:
  - chains.json   (Chiliz Mainnet + Spicy Testnet)
  - tokens.json   (PEPPER token)
  - contacts.json (empty address book)
  - .env.example  (wallet key template)

After init, run:
  chiliz chain-info                    # verify connection
  chiliz address 0x...                 # query any address
  chiliz validators                    # Chiliz validator set
  chiliz --chain "spicy" chain-info    # use testnet`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot find home directory: %w", err)
		}

		chilizDir := filepath.Join(home, ".chiliz")

		// Create directory
		if err := os.MkdirAll(chilizDir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", chilizDir, err)
		}

		// Write config files
		files := []struct {
			name string
			data interface{}
		}{
			{"chains.json", defaultChains},
			{"tokens.json", defaultTokens},
			{"contacts.json", defaultContacts},
		}

		for _, f := range files {
			path := filepath.Join(chilizDir, f.name)
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("  [skip] %s already exists\n", path)
				continue
			}

			content, err := json.MarshalIndent(f.data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal %s: %w", f.name, err)
			}

			if err := os.WriteFile(path, append(content, '\n'), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
			fmt.Printf("  [created] %s\n", path)
		}

		// Write .env.example
		envPath := filepath.Join(chilizDir, ".env.example")
		if _, err := os.Stat(envPath); err != nil {
			envContent := "# Wallet Private Keys (without 0x prefix)\nCHILIZ_WALLET_MAIN=\nCHILIZ_WALLET_TEST=\n"
			if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
				return fmt.Errorf("failed to write .env.example: %w", err)
			}
			fmt.Printf("  [created] %s\n", envPath)
		} else {
			fmt.Printf("  [skip] %s already exists\n", envPath)
		}

		fmt.Println()
		fmt.Println("  Chiliz initialized! Try:")
		fmt.Println("    chiliz chain-info")
		fmt.Println("    chiliz validators")
		fmt.Printf("    chiliz --chain \"spicy\" chain-info   # testnet\n")
		fmt.Println()

		return nil
	},
}
