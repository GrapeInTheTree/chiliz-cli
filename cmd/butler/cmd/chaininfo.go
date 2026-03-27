package cmd

import (
	"sync"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/ethereum"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/output"
	"github.com/spf13/cobra"
)

var chainInfoCmd = &cobra.Command{
	Use:   "chain-info",
	Short: "Show chain status",
	Long:  "Display current chain information including latest block and gas price",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		chain := appCtx.Chain
		rpc := chain.RPCURL

		var (
			latestBlock uint64
			gasPrice    string
			wg          sync.WaitGroup
			mu          sync.Mutex
		)

		wg.Add(2)

		go func() {
			defer wg.Done()
			n, err := ethereum.GetLatestBlockNumber(rpc)
			if err == nil {
				mu.Lock()
				latestBlock = n
				mu.Unlock()
			}
		}()

		go func() {
			defer wg.Done()
			gp, err := ethereum.GetGasPrice(rpc)
			if err == nil {
				mu.Lock()
				gasPrice = formatGwei(gp)
				mu.Unlock()
			}
		}()

		wg.Wait()

		info := domain.ChainStatus{
			Name:        chain.Name,
			ChainID:     chain.ChainID,
			RPCURL:      chain.RPCURL,
			LatestBlock: latestBlock,
			GasPrice:    gasPrice,
			Currency:    chain.CurrencySymbol,
		}

		return output.Print(jsonOutput, info)
	},
}
