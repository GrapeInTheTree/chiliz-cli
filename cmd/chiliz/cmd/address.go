package cmd

import (
	"fmt"
	"sync"

	"github.com/GrapeInTheTree/chiliz-cli/internal/domain"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/config"
	"github.com/GrapeInTheTree/chiliz-cli/internal/infra/ethereum"
	"github.com/GrapeInTheTree/chiliz-cli/internal/output"
	"github.com/spf13/cobra"
)

var addressCmd = &cobra.Command{
	Use:   "address <addr>",
	Short: "Balance, nonce, tx history, and token holdings for an address",
	Long:  "Display balance, nonce, token holdings, and recent transactions for an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := config.ResolveAddress(args[0], appCtx.Contacts)
		if err != nil {
			return err
		}

		chain := appCtx.Chain
		rpc := chain.RPCURL

		// Parallel RPC fetches
		var (
			balance    string
			nonce      uint64
			isContract bool
			rpcErr     error
			mu         sync.Mutex
			wg         sync.WaitGroup
		)

		wg.Add(3)

		go func() {
			defer wg.Done()
			bal, err := ethereum.GetBalance(rpc, addr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				rpcErr = fmt.Errorf("balance: %w", err)
				return
			}
			balance = ethereum.FormatBalance(bal)
		}()

		go func() {
			defer wg.Done()
			n, err := ethereum.GetNonce(rpc, addr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				rpcErr = fmt.Errorf("nonce: %w", err)
				return
			}
			nonce = n
		}()

		go func() {
			defer wg.Done()
			code, err := ethereum.GetCode(rpc, addr)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				return // non-fatal
			}
			isContract = len(code) > 0
		}()

		wg.Wait()
		if rpcErr != nil {
			return rpcErr
		}

		info := domain.AddressInfo{
			Address:       addr,
			Chain:         chain.Name,
			ChainID:       chain.ChainID,
			NativeBalance: balance,
			NativeSymbol:  chain.CurrencySymbol,
			Nonce:         nonce,
			IsContract:    isContract,
		}

		// Explorer data (graceful degradation)
		if appCtx.Explorer != nil {
			var explWg sync.WaitGroup
			explWg.Add(3)

			go func() {
				defer explWg.Done()
				txs, err := appCtx.Explorer.GetTxList(addr, 1, 10)
				if err == nil {
					mu.Lock()
					info.RecentTxs = txs
					mu.Unlock()
				}
			}()

			go func() {
				defer explWg.Done()
				tokens, err := appCtx.Explorer.GetTokenBalances(addr)
				if err == nil {
					mu.Lock()
					info.TokenBalances = tokens
					mu.Unlock()
				}
			}()

			go func() {
				defer explWg.Done()
				itxs, err := appCtx.Explorer.GetInternalTxList(addr, 1, 5)
				if err == nil {
					mu.Lock()
					info.InternalTxs = itxs
					mu.Unlock()
				}
			}()

			explWg.Wait()
		}

		return output.Print(jsonOutput, info)
	},
}
