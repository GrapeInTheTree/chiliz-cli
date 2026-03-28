package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/GrapeInTheTree/chiliz-cli/internal/domain"
	"github.com/joho/godotenv"
)

// configDir holds the resolved configuration directory
var configDir string

// SetConfigDir sets the config directory explicitly
func SetConfigDir(dir string) {
	configDir = dir
}

// ResolveConfigDir determines the config directory using a cascade:
// 1. explicit flag (if non-empty)
// 2. CHILIZ_CONFIG_DIR env var
// 3. ~/.chiliz/ (if chains.json exists there)
// 4. current working directory
func ResolveConfigDir(flagValue string) string {
	if flagValue != "" {
		configDir = flagValue
		return configDir
	}

	if envDir := os.Getenv("CHILIZ_CONFIG_DIR"); envDir != "" {
		configDir = envDir
		return configDir
	}

	if home, err := os.UserHomeDir(); err == nil {
		chilizHome := filepath.Join(home, ".chiliz")
		if _, err := os.Stat(filepath.Join(chilizHome, "chains.json")); err == nil {
			configDir = chilizHome
			return configDir
		}
	}

	// Fallback: current working directory (preserves existing behavior)
	configDir = ""
	return configDir
}

// configPath returns the full path to a config file
func configPath(filename string) string {
	if configDir == "" {
		return filename
	}
	return filepath.Join(configDir, filename)
}

// LoadChains loads blockchain configurations from chains.json
func LoadChains() ([]domain.Chain, error) {
	data, err := os.ReadFile(configPath("chains.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read chains.json: %w", err)
	}

	var chains []domain.Chain
	if err := json.Unmarshal(data, &chains); err != nil {
		return nil, fmt.Errorf("failed to parse chains.json: %w", err)
	}

	slog.Info("Loaded chains", "count", len(chains))
	return chains, nil
}

// LoadContacts loads address book from contacts.json
func LoadContacts() ([]domain.Contact, error) {
	data, err := os.ReadFile(configPath("contacts.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read contacts.json: %w", err)
	}

	var contacts []domain.Contact
	if err := json.Unmarshal(data, &contacts); err != nil {
		return nil, fmt.Errorf("failed to parse contacts.json: %w", err)
	}

	slog.Info("Loaded contacts", "count", len(contacts))
	return contacts, nil
}

// LoadWallets loads wallet configuration and attempts to load .env
func LoadWallets() ([]domain.Wallet, error) {
	// Try to load .env file from config dir, then from cwd
	envPath := configPath(".env")
	if err := godotenv.Load(envPath); err != nil {
		// Also try current directory if configDir is set
		if configDir != "" {
			_ = godotenv.Load(".env")
		}
		slog.Warn("No .env file found - please create one from .env.example")
	}

	// Hardcoded wallet list (only names and env keys)
	wallets := []domain.Wallet{
		{Name: "Main Wallet", EnvKey: "CHILIZ_WALLET_MAIN"},
		{Name: "Test Wallet", EnvKey: "CHILIZ_WALLET_TEST"},
	}

	slog.Info("Loaded wallet configurations", "count", len(wallets))
	return wallets, nil
}

// LoadTokens loads ERC-20 token configurations from tokens.json
func LoadTokens() ([]domain.Token, error) {
	data, err := os.ReadFile(configPath("tokens.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read tokens.json: %w", err)
	}

	var tokens []domain.Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse tokens.json: %w", err)
	}

	slog.Info("Loaded tokens", "count", len(tokens))
	return tokens, nil
}

// GetTokensForChain returns all tokens for a specific chain ID
func GetTokensForChain(tokens []domain.Token, chainID int64) []domain.Token {
	var result []domain.Token
	for _, token := range tokens {
		if token.ChainID == chainID {
			result = append(result, token)
		}
	}
	return result
}

// ResolveAddress resolves a contact name or validates an Ethereum address.
// If input starts with "0x", it validates as an address (42 chars).
// Otherwise, it searches contacts by name (case-insensitive partial match).
func ResolveAddress(input string, contacts []domain.Contact) (string, error) {
	if strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "0X") {
		if len(input) != 42 {
			return "", fmt.Errorf("invalid address: must be 0x + 40 hex chars")
		}
		return input, nil
	}

	// Search contacts by name (case-insensitive, partial match)
	inputLower := strings.ToLower(input)
	for _, c := range contacts {
		if strings.Contains(strings.ToLower(c.Name), inputLower) {
			return c.Address, nil
		}
	}

	return "", fmt.Errorf("address or contact %q not found\nHint: use a 0x address or a name from contacts.json", input)
}

// GetPrivateKey safely retrieves a private key from environment
func GetPrivateKey(envKey string) (string, error) {
	key := os.Getenv(envKey)
	if key == "" {
		return "", fmt.Errorf("private key not found for %s", envKey)
	}
	return key, nil
}
