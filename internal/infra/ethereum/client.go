package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetBalance retrieves the balance of an address on a specific chain
func GetBalance(rpcURL, address string) (*big.Int, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	addr := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	slog.Info("Balance retrieved", "address", address, "balance", balance.String())
	return balance, nil
}

// SendTransaction sends native currency (ETH/CHZ/etc) from one address to another
func SendTransaction(
	privateKeyHex string,
	rpcURL string,
	chainID *big.Int,
	toAddress string,
	amount *big.Int,
) (string, error) {
	// 1. Connect to RPC
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	// 2. Load private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// 3. Derive sender address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("failed to cast public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 4. Get nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// 5. Get gas price suggestions (EIP-1559)
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

	// 6. Set gas limit for simple transfer
	gasLimit := uint64(21000)

	// 7. Create transaction
	toAddr := common.HexToAddress(toAddress)
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddr,
		Value:     amount,
		Data:      nil,
	})

	// 8. Sign transaction
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 9. Send transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	slog.Info("Transaction sent successfully",
		"hash", txHash,
		"from", fromAddress.Hex(),
		"to", toAddress,
		"amount", amount.String())

	return txHash, nil
}

// FormatBalance converts Wei to Ether/CHZ (with 18 decimals)
func FormatBalance(balance *big.Int) string {
	fBalance := new(big.Float).SetInt(balance)
	ethValue := new(big.Float).Quo(fBalance, big.NewFloat(1e18))
	return ethValue.Text('f', 6)
}

// ParseAmount converts string amount to Wei (*big.Int)
func ParseAmount(amountStr string) (*big.Int, error) {
	fAmount, _, err := big.ParseFloat(amountStr, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	// Multiply by 10^18 to convert to Wei
	weiFloat := new(big.Float).Mul(fAmount, big.NewFloat(1e18))
	wei, _ := weiFloat.Int(nil)

	return wei, nil
}

// GetAddressFromPrivateKey derives the Ethereum address from a private key hex string
func GetAddressFromPrivateKey(privateKeyHex string) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("failed to cast public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex(), nil
}
