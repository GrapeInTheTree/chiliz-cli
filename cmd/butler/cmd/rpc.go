package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var rpcCmd = &cobra.Command{
	Use:   "rpc <method> [params]",
	Short: "Raw JSON-RPC call (any method, e.g., eth_feeHistory)",
	Long: `Send an arbitrary JSON-RPC call to the chain's RPC endpoint.
Params should be a JSON array string. If omitted, empty array is used.

Examples:
  butler rpc eth_blockNumber
  butler rpc eth_getBalance '["0xC3B2...", "latest"]'
  butler rpc eth_feeHistory '["4", "latest", []]'`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := args[0]
		params := "[]"
		if len(args) > 1 {
			params = args[1]
		}

		// Validate params is valid JSON
		if !json.Valid([]byte(params)) {
			return fmt.Errorf("invalid JSON params: %s", params)
		}

		// Build JSON-RPC request
		reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":%s,"id":1}`, method, params)

		resp, err := http.Post(appCtx.Chain.RPCURL, "application/json", bytes.NewBufferString(reqBody))
		if err != nil {
			return fmt.Errorf("RPC request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		// Pretty-print the JSON response
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			// If not valid JSON, print raw
			fmt.Println(string(body))
			return nil
		}

		prettyJSON.WriteTo(os.Stdout)
		fmt.Println()
		return nil
	},
}
