package cmd

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/ethereum"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/output"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const stakingContract = "0x0000000000000000000000000000000000001000"

var validatorsCmd = &cobra.Command{
	Use:   "validators",
	Short: "Chiliz validators: status, delegated, commission, rewards",
	Long:  "Query the Staking system contract (0x...1000) for active validators, delegated amounts, and commission rates.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rpc := appCtx.Chain.RPCURL

		// 1. Get validator list
		calldata, err := ethereum.BuildCalldata("getValidators()", nil)
		if err != nil {
			return fmt.Errorf("encode getValidators: %w", err)
		}

		resultBytes, err := ethereum.CallContract(rpc, stakingContract, calldata)
		if err != nil {
			return fmt.Errorf("getValidators call failed: %w", err)
		}

		decoded, err := ethereum.DecodeOutputs("(address[])", resultBytes)
		if err != nil {
			return fmt.Errorf("decode validators: %w", err)
		}

		// Parse addresses from the decoded string "[0x..., 0x..., ...]"
		addresses := parseAddressArray(resultBytes)
		if len(addresses) == 0 {
			return fmt.Errorf("no validators found")
		}

		// 2. Get status for each validator in parallel
		validators := make([]domain.ValidatorInfo, len(addresses))
		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstErr error

		for i, addr := range addresses {
			wg.Add(1)
			go func(idx int, validatorAddr string) {
				defer wg.Done()

				info, err := fetchValidatorStatus(rpc, validatorAddr)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					if firstErr == nil {
						firstErr = fmt.Errorf("validator %s: %w", validatorAddr, err)
					}
					// Still fill basic info
					validators[idx] = domain.ValidatorInfo{
						Address: validatorAddr,
						Status:  "unknown",
					}
					return
				}
				info.Address = validatorAddr
				validators[idx] = info
			}(i, addr)
		}
		wg.Wait()

		_ = decoded // already used via parseAddressArray

		result := domain.ValidatorsResult{
			Chain:      appCtx.Chain.Name,
			ChainID:    appCtx.Chain.ChainID,
			Count:      len(validators),
			Validators: validators,
		}

		return output.Print(jsonOutput, result)
	},
}

// fetchValidatorStatus queries getValidatorStatus for a single validator
func fetchValidatorStatus(rpcURL, validatorAddr string) (domain.ValidatorInfo, error) {
	calldata, err := ethereum.BuildCalldata("getValidatorStatus(address)", []string{validatorAddr})
	if err != nil {
		return domain.ValidatorInfo{}, err
	}

	resultBytes, err := ethereum.CallContract(rpcURL, stakingContract, calldata)
	if err != nil {
		return domain.ValidatorInfo{}, err
	}

	// getValidatorStatus returns:
	// (address ownerAddress, uint8 status, uint256 totalDelegated, uint32 slashesCount,
	//  uint64 changedAt, uint64 jailedBefore, uint64 claimedAt, uint16 commissionRate, uint96 totalRewards)
	values, err := ethereum.DecodeOutputs(
		"(address,uint8,uint256,uint32,uint64,uint64,uint64,uint16,uint96)",
		resultBytes,
	)
	if err != nil {
		return domain.ValidatorInfo{}, err
	}

	if len(values) < 9 {
		return domain.ValidatorInfo{}, fmt.Errorf("unexpected return values: got %d", len(values))
	}

	// Format delegated amount (wei → CHZ)
	delegated := new(big.Int)
	delegated.SetString(values[2], 10)
	formattedDelegated := ethereum.FormatBalance(delegated) + " CHZ"

	// Format rewards (wei → CHZ)
	rewards := new(big.Int)
	rewards.SetString(values[8], 10)
	formattedRewards := ethereum.FormatBalance(rewards) + " CHZ"

	// Commission rate: value is basis points (e.g., 100 = 1.0%)
	commissionBP := values[7]
	commission := new(big.Float)
	commission.SetString(commissionBP)
	commissionPct := new(big.Float).Quo(commission, big.NewFloat(100))
	commissionStr := commissionPct.Text('f', 1) + "%"

	return domain.ValidatorInfo{
		Owner:          values[0],
		Status:         statusName(values[1]),
		TotalDelegated: formattedDelegated,
		SlashCount:     parseUint32(values[3]),
		CommissionRate: commissionStr,
		TotalRewards:   formattedRewards,
	}, nil
}

// parseAddressArray decodes ABI-encoded address[] from raw bytes
func parseAddressArray(data []byte) []string {
	// Use abi_helper to decode
	values, err := ethereum.DecodeOutputs("(address[])", data)
	if err != nil || len(values) == 0 {
		return nil
	}

	// values[0] is "[0x..., 0x..., ...]" — parse by re-decoding with abi
	// Actually, let's use the raw ABI decoding directly
	args, err := ethereum.DecodeRawOutputs("(address[])", data)
	if err != nil || len(args) == 0 {
		return nil
	}

	// args[0] should be []common.Address
	addrs, ok := args[0].([]common.Address)
	if !ok {
		return nil
	}

	result := make([]string, len(addrs))
	for i, a := range addrs {
		result[i] = a.Hex()
	}
	return result
}

func statusName(s string) string {
	switch s {
	case "0":
		return "NotFound"
	case "1":
		return "Active"
	case "2":
		return "Pending"
	case "3":
		return "Jail"
	default:
		return "Unknown"
	}
}

func parseUint32(s string) uint32 {
	n := new(big.Int)
	n.SetString(s, 10)
	return uint32(n.Uint64())
}
