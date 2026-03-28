package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/config"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/infra/ethereum"
	"github.com/GrapeInTheTree/go-ethereum-butler/internal/output"
	"github.com/spf13/cobra"
)

var callCmd = &cobra.Command{
	Use:   "call <contract> <signature> [args...]",
	Short: "Read-only contract call via eth_call (like cast call)",
	Long: `Execute a read-only eth_call against a smart contract.

Signature format: "functionName(inputTypes)(outputTypes)"
Output types are optional — omit them to get raw hex.

Examples:
  butler call 0x60F3...6b67 "totalSupply()(uint256)"
  butler call 0x0...1000 "isValidator(address)(bool)" 0x8d9b...
  butler call 0x0...1000 "getValidators()(address[])"
  butler call 0x60F3...6b67 "totalSupply()"              # raw hex output`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		contractAddr, err := config.ResolveAddress(args[0], appCtx.Contacts)
		if err != nil {
			return err
		}
		sig := args[1]
		callArgs := args[2:]

		// Parse signature into input and output parts
		inputSig, outputTypes, err := ethereum.ParseCallSignature(sig)
		if err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}

		// Build calldata (4-byte selector + ABI-encoded args)
		calldata, err := ethereum.BuildCalldata(inputSig, callArgs)
		if err != nil {
			return fmt.Errorf("failed to encode call: %w", err)
		}

		// Execute eth_call
		resultBytes, err := ethereum.CallContract(appCtx.Chain.RPCURL, contractAddr, calldata)
		if err != nil {
			return err
		}

		rawHex := "0x" + hex.EncodeToString(resultBytes)

		// Decode return values if output types were specified
		var values []string
		if outputTypes != "" {
			values, err = ethereum.DecodeOutputs(outputTypes, resultBytes)
			if err != nil {
				return fmt.Errorf("failed to decode output: %w", err)
			}
		}

		result := domain.CallResult{
			Contract: contractAddr,
			Method:   inputSig,
			Values:   values,
			Raw:      rawHex,
		}

		return output.Print(jsonOutput, result)
	},
}
