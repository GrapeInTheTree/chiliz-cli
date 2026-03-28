package cmd

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/ethereum"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/output"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const stakingPoolContract = "0x0000000000000000000000000000000000007001"

var stakingCmd = &cobra.Command{
	Use:   "staking <address>",
	Short: "Show staking positions for an address",
	Long:  "Query StakingPool (0x...7001) for staked amounts and claimable rewards per validator.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := args[0]
		if !strings.HasPrefix(addr, "0x") || len(addr) != 42 {
			return fmt.Errorf("invalid address format: must be 0x + 40 hex chars")
		}

		rpc := appCtx.Chain.RPCURL

		// 1. Get validator list from Staking contract
		calldata, err := ethereum.BuildCalldata("getValidators()", nil)
		if err != nil {
			return err
		}
		resultBytes, err := ethereum.CallContract(rpc, stakingContract, calldata)
		if err != nil {
			return err
		}

		rawValues, err := ethereum.DecodeRawOutputs("(address[])", resultBytes)
		if err != nil {
			return err
		}
		validators, ok := rawValues[0].([]common.Address)
		if !ok || len(validators) == 0 {
			return fmt.Errorf("no validators found")
		}

		// 2. Query staked amount + rewards per validator in parallel
		type stakingResult struct {
			validator string
			staked    *big.Int
			rewards   *big.Int
		}

		results := make([]stakingResult, len(validators))
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i, v := range validators {
			wg.Add(1)
			go func(idx int, validatorAddr string) {
				defer wg.Done()

				staked := queryStakingPool(rpc, "getStakedAmount(address,address)", validatorAddr, addr)
				rewards := queryStakingPool(rpc, "claimableRewards(address,address)", validatorAddr, addr)

				mu.Lock()
				results[idx] = stakingResult{
					validator: validatorAddr,
					staked:    staked,
					rewards:   rewards,
				}
				mu.Unlock()
			}(i, v.Hex())
		}
		wg.Wait()

		// 3. Filter validators where user has staked > 0 and build output
		totalStaked := new(big.Int)
		totalRewards := new(big.Int)
		var entries []domain.StakingEntry

		for _, r := range results {
			if r.staked != nil && r.staked.Sign() > 0 {
				totalStaked.Add(totalStaked, r.staked)
				if r.rewards != nil {
					totalRewards.Add(totalRewards, r.rewards)
				}
				entries = append(entries, domain.StakingEntry{
					Validator: r.validator,
					Staked:    ethereum.FormatBalance(r.staked) + " CHZ",
					Rewards:   ethereum.FormatBalance(r.rewards) + " CHZ",
				})
			}
		}

		info := domain.StakingInfo{
			Address:      addr,
			Chain:        appCtx.Chain.Name,
			TotalStaked:  ethereum.FormatBalance(totalStaked) + " CHZ",
			TotalRewards: ethereum.FormatBalance(totalRewards) + " CHZ",
			Entries:      entries,
		}

		return output.Print(jsonOutput, info)
	},
}

// queryStakingPool calls a 2-arg function on StakingPool contract and returns the uint256 result.
func queryStakingPool(rpc, sig, validatorAddr, stakerAddr string) *big.Int {
	calldata, err := ethereum.BuildCalldata(sig, []string{validatorAddr, stakerAddr})
	if err != nil {
		return big.NewInt(0)
	}

	resultBytes, err := ethereum.CallContract(rpc, stakingPoolContract, calldata)
	if err != nil {
		return big.NewInt(0)
	}

	values, err := ethereum.DecodeRawOutputs("(uint256)", resultBytes)
	if err != nil || len(values) == 0 {
		return big.NewInt(0)
	}

	n, ok := values[0].(*big.Int)
	if !ok {
		return big.NewInt(0)
	}
	return n
}
