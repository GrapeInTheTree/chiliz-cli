package cmd

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/ethereum"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/output"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
)

var txCmd = &cobra.Command{
	Use:   "tx <hash>",
	Short: "Show transaction details",
	Long:  "Display full details for a transaction by hash",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash := args[0]
		if !strings.HasPrefix(hash, "0x") || len(hash) != 66 {
			return fmt.Errorf("invalid tx hash format: must be 0x + 64 hex chars")
		}

		chain := appCtx.Chain
		rpc := chain.RPCURL

		var (
			tx        *types.Transaction
			receipt   *types.Receipt
			isPending bool
			txErr     error
			rcptErr   error
			wg        sync.WaitGroup
		)

		wg.Add(2)

		go func() {
			defer wg.Done()
			tx, isPending, txErr = ethereum.GetTransaction(rpc, hash)
		}()

		go func() {
			defer wg.Done()
			receipt, rcptErr = ethereum.GetTransactionReceipt(rpc, hash)
		}()

		wg.Wait()

		if txErr != nil {
			return fmt.Errorf("failed to get transaction: %w", txErr)
		}

		// Derive sender
		signer := types.LatestSignerForChainID(chain.GetChainIDBigInt())
		from, _ := types.Sender(signer, tx)

		to := ""
		if tx.To() != nil {
			to = tx.To().Hex()
		}

		// Format value
		valueFormatted := ethereum.FormatBalance(tx.Value()) + " " + chain.CurrencySymbol

		// Gas price in Gwei
		gasPrice := formatGwei(tx.GasPrice())

		// Status
		status := "pending"
		var blockNumber uint64
		var gasUsed uint64
		var txFee string
		var timestamp int64
		var timeHuman string
		var logCount int

		if !isPending && rcptErr == nil && receipt != nil {
			if receipt.Status == 1 {
				status = "success"
			} else {
				status = "failed"
			}
			blockNumber = receipt.BlockNumber.Uint64()
			gasUsed = receipt.GasUsed
			logCount = len(receipt.Logs)

			// Calculate tx fee
			fee := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), tx.GasPrice())
			txFee = ethereum.FormatBalance(fee) + " " + chain.CurrencySymbol

			// Get block timestamp
			block, err := ethereum.GetBlock(rpc, receipt.BlockNumber)
			if err == nil {
				timestamp = int64(block.Time())
				timeHuman = time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05 UTC")
			}
		}

		// Method ID
		methodID := ""
		inputData := "0x"
		if len(tx.Data()) > 0 {
			inputData = "0x" + hex.EncodeToString(tx.Data())
			if len(tx.Data()) >= 4 {
				methodID = "0x" + hex.EncodeToString(tx.Data()[:4])
			}
		}

		detail := domain.TxDetail{
			Hash:           hash,
			Status:         status,
			BlockNumber:    blockNumber,
			Timestamp:      timestamp,
			TimeHuman:      timeHuman,
			From:           from.Hex(),
			To:             to,
			Value:          tx.Value().String(),
			ValueFormatted: valueFormatted,
			GasPrice:       gasPrice,
			GasUsed:        gasUsed,
			GasLimit:       tx.Gas(),
			TxFee:          txFee,
			Nonce:          tx.Nonce(),
			InputData:      inputData,
			MethodID:       methodID,
			LogCount:       logCount,
		}

		return output.Print(jsonOutput, detail)
	},
}

// formatGwei converts a gas price (wei) to Gwei string
func formatGwei(wei *big.Int) string {
	if wei == nil {
		return "0"
	}
	gwei := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		new(big.Float).SetFloat64(1e9),
	)
	return gwei.Text('f', 2) + " Gwei"
}
