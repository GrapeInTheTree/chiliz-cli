package ethereum

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/ethereum/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetTokenBalance retrieves the ERC-20 token balance for an address using generated bindings
func GetTokenBalance(rpcURL, tokenAddress, walletAddress string, decimals int) (*big.Int, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	// Create contract instance using generated binding
	tokenAddr := common.HexToAddress(tokenAddress)
	instance, err := contracts.NewERC20(tokenAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create token instance: %w", err)
	}

	// Get balance
	walletAddr := common.HexToAddress(walletAddress)
	balance, err := instance.BalanceOf(&bind.CallOpts{}, walletAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get token balance: %w", err)
	}

	return balance, nil
}

// SendTokenTransaction sends ERC-20 tokens from one address to another using generated bindings
func SendTokenTransaction(
	privateKeyHex string,
	rpcURL string,
	chainID *big.Int,
	tokenAddress string,
	toAddress string,
	amount *big.Int,
) (string, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	// Load private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// Get gas prices
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas tip cap: %w", err)
	}

	gasFeeCap, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas fee cap: %w", err)
	}

	// Add buffer to gasFeeCap
	gasFeeCap = new(big.Int).Mul(gasFeeCap, big.NewInt(12))
	gasFeeCap = new(big.Int).Div(gasFeeCap, big.NewInt(10)) // 1.2x

	// Create auth options
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return "", fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.GasTipCap = gasTipCap
	auth.GasFeeCap = gasFeeCap
	// auth.GasLimit will be automatically estimated by bind

	// Create contract instance
	tokenAddr := common.HexToAddress(tokenAddress)
	instance, err := contracts.NewERC20(tokenAddr, client)
	if err != nil {
		return "", fmt.Errorf("failed to create token instance: %w", err)
	}

	// Send transaction
	toAddr := common.HexToAddress(toAddress)
	tx, err := instance.Transfer(auth, toAddr, amount)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := tx.Hash().Hex()
	slog.Info("Token transaction sent successfully",
		"hash", txHash,
		"token", tokenAddress,
		"amount", amount.String())

	return txHash, nil
}

// FormatTokenBalance converts token units to human-readable format based on decimals
func FormatTokenBalance(balance *big.Int, decimals int) string {
	fBalance := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetInt(pow10(decimals))
	tokenValue := new(big.Float).Quo(fBalance, divisor)
	return tokenValue.Text('f', 6)
}

// ParseTokenAmount converts string amount to token units based on decimals
func ParseTokenAmount(amountStr string, decimals int) (*big.Int, error) {
	fAmount, _, err := big.ParseFloat(amountStr, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	// Multiply by 10^decimals
	multiplier := new(big.Float).SetInt(pow10(decimals))
	tokenUnits := new(big.Float).Mul(fAmount, multiplier)
	result, _ := tokenUnits.Int(nil)

	return result, nil
}

// pow10 calculates 10^n as *big.Int (safe for any decimal count)
func pow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}
