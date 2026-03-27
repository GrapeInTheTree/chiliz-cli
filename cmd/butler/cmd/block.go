package cmd

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/ethereum"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/output"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block [number|latest]",
	Short: "Show block information",
	Long:  "Display block details by number. Defaults to latest block.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rpc := appCtx.Chain.RPCURL

		var blockNum *big.Int // nil = latest
		if len(args) > 0 && args[0] != "latest" {
			n, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid block number: %s", args[0])
			}
			blockNum = new(big.Int).SetUint64(n)
		}

		block, err := ethereum.GetBlock(rpc, blockNum)
		if err != nil {
			return err
		}

		baseFee := ""
		if block.BaseFee() != nil {
			baseFee = formatGwei(block.BaseFee())
		}

		info := domain.BlockInfo{
			Number:     block.NumberU64(),
			Hash:       block.Hash().Hex(),
			ParentHash: block.ParentHash().Hex(),
			Timestamp:  int64(block.Time()),
			TimeHuman:  time.Unix(int64(block.Time()), 0).UTC().Format("2006-01-02 15:04:05 UTC"),
			GasUsed:    block.GasUsed(),
			GasLimit:   block.GasLimit(),
			TxCount:    len(block.Transactions()),
			Miner:      block.Coinbase().Hex(),
			BaseFee:    baseFee,
		}

		return output.Print(jsonOutput, info)
	},
}
