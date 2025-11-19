package domain

import "math/big"

// Chain represents an EVM-compatible blockchain network
type Chain struct {
	Name           string `json:"name"`
	RPCURL         string `json:"rpc_url"`
	ChainID        int64  `json:"chain_id"`
	CurrencySymbol string `json:"currency_symbol"`
	LogoURL        string `json:"logo_url"`
}

// GetChainIDBigInt returns the chain ID as *big.Int
func (c *Chain) GetChainIDBigInt() *big.Int {
	return big.NewInt(c.ChainID)
}

// Contact represents an address book entry
type Contact struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// Token represents an ERC-20 token on a specific chain
type Token struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
	ChainID  int64  `json:"chain_id"`
	LogoURL  string `json:"logo_url"`
}

// IsNative returns true if this is a placeholder for native token
func (t *Token) IsNative() bool {
	return t.Address == "" || t.Address == "0x0000000000000000000000000000000000000000"
}

// Wallet represents a wallet configuration
type Wallet struct {
	Name   string
	EnvKey string
}
