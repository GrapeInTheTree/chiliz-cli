package cmd

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/GrapeInTheTree/chiliz-cli/internal/domain"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/ethereum"
	"github.com/GrapeInTheTree/chiliz-cli/internal/output"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var (
	logsAddress   string
	logsEvent     string
	logsBlocks    uint64
	logsFromBlock uint64
	logsToBlock   uint64
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Filter on-chain event logs via eth_getLogs (RPC)",
	Long: `Query event logs directly from the chain via RPC.

Examples:
  butler logs --address 0x60F3...6b67 --event "Transfer(address,address,uint256)" --blocks 500
  butler logs --address 0x0...1000 --blocks 100
  butler logs --address 0x60F3... --from-block 32640000 --to-block 32641000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if logsAddress == "" {
			return fmt.Errorf("--address is required")
		}
		if !strings.HasPrefix(logsAddress, "0x") || len(logsAddress) != 42 {
			return fmt.Errorf("invalid address: must be 0x + 40 hex chars")
		}

		rpc := appCtx.Chain.RPCURL

		// Determine block range
		var fromBlock, toBlock *big.Int

		if logsFromBlock > 0 || logsToBlock > 0 {
			if logsFromBlock > 0 {
				fromBlock = new(big.Int).SetUint64(logsFromBlock)
			}
			if logsToBlock > 0 {
				toBlock = new(big.Int).SetUint64(logsToBlock)
			}
		} else {
			// Use --blocks flag (default 1000)
			latest, err := ethereum.GetLatestBlockNumber(rpc)
			if err != nil {
				return fmt.Errorf("failed to get latest block: %w", err)
			}
			blocks := logsBlocks
			if blocks == 0 {
				blocks = 1000
			}
			if latest > blocks {
				fromBlock = new(big.Int).SetUint64(latest - blocks)
			} else {
				fromBlock = big.NewInt(0)
			}
		}

		// Build topic filter
		var topics []common.Hash
		if logsEvent != "" {
			topicHash := crypto.Keccak256Hash([]byte(logsEvent))
			topics = append(topics, topicHash)
		}

		logs, err := ethereum.FilterLogs(rpc, logsAddress, topics, fromBlock, toBlock)
		if err != nil {
			return fmt.Errorf("failed to query logs: %w", err)
		}

		// Convert to domain types
		entries := make([]domain.LogEntry, len(logs))
		for i, l := range logs {
			topicStrs := make([]string, len(l.Topics))
			for j, t := range l.Topics {
				topicStrs[j] = t.Hex()
			}

			eventName := ""
			if logsEvent != "" {
				eventName = logsEvent
			} else if len(l.Topics) > 0 {
				// Try to resolve event name via 4byte
				eventName = ethereum.LookupEventTopic(l.Topics[0].Hex())
			}

			entries[i] = domain.LogEntry{
				Address:     l.Address.Hex(),
				BlockNumber: l.BlockNumber,
				TxHash:      l.TxHash.Hex(),
				Topics:      topicStrs,
				Data:        "0x" + hex.EncodeToString(l.Data),
				EventName:   eventName,
			}
		}

		result := domain.LogsResult{
			Address: logsAddress,
			Event:   logsEvent,
			Count:   len(entries),
			Logs:    entries,
		}

		return output.Print(jsonOutput, result)
	},
}

func init() {
	logsCmd.Flags().StringVar(&logsAddress, "address", "", "contract address to filter logs")
	logsCmd.Flags().StringVar(&logsEvent, "event", "", "event signature (e.g., Transfer(address,address,uint256))")
	logsCmd.Flags().Uint64Var(&logsBlocks, "blocks", 1000, "number of recent blocks to scan")
	logsCmd.Flags().Uint64Var(&logsFromBlock, "from-block", 0, "start block number")
	logsCmd.Flags().Uint64Var(&logsToBlock, "to-block", 0, "end block number")
}
