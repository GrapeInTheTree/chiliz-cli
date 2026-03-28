package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
)

// Print outputs the value in JSON or human-readable format
func Print(jsonMode bool, v any) error {
	if jsonMode {
		return printJSON(v)
	}
	return printHuman(v)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printHuman(v any) error {
	switch data := v.(type) {
	case domain.AddressInfo:
		printAddressHuman(data)
	case domain.TxDetail:
		printTxHuman(data)
	case domain.BlockInfo:
		printBlockHuman(data)
	case domain.ChainStatus:
		printChainInfoHuman(data)
	case domain.CallResult:
		printCallHuman(data)
	case domain.ValidatorsResult:
		printValidatorsHuman(data)
	case domain.StakingInfo:
		printStakingHuman(data)
	default:
		return printJSON(v)
	}
	return nil
}

func printAddressHuman(a domain.AddressInfo) {
	fmt.Println()
	fmt.Printf("  Address:  %s\n", a.Address)
	fmt.Printf("  Chain:    %s (%d)\n", a.Chain, a.ChainID)
	fmt.Printf("  Balance:  %s %s\n", a.NativeBalance, a.NativeSymbol)
	fmt.Printf("  Nonce:    %d\n", a.Nonce)
	if a.IsContract {
		fmt.Println("  Type:     Contract")
	} else {
		fmt.Println("  Type:     EOA")
	}

	if len(a.TokenBalances) > 0 {
		fmt.Println()
		fmt.Println("  Token Holdings:")
		for _, t := range a.TokenBalances {
			fmt.Printf("    %-10s %s\n", t.Symbol, t.Balance)
		}
	}

	if len(a.RecentTxs) > 0 {
		fmt.Println()
		fmt.Printf("  Recent Transactions (last %d):\n", len(a.RecentTxs))
		fmt.Printf("    %-15s %-12s %-18s %s\n", "Hash", "Method", "Value", "Time")
		fmt.Printf("    %s\n", strings.Repeat("-", 65))
		for _, tx := range a.RecentTxs {
			hash := shortenHash(tx.Hash)
			method := tx.Method
			if len(method) > 10 {
				method = method[:10] + ".."
			}

			direction := ""
			addrLower := strings.ToLower(a.Address)
			if strings.ToLower(tx.To) == addrLower {
				direction = fmt.Sprintf("+%s", tx.Value)
			} else {
				direction = fmt.Sprintf("-%s", tx.Value)
			}
			if len(direction) > 16 {
				direction = direction[:16] + ".."
			}

			fmt.Printf("    %-15s %-12s %-18s %s\n", hash, method, direction, relativeTime(tx.Timestamp))
		}
	}
	fmt.Println()
}

func printTxHuman(t domain.TxDetail) {
	fmt.Println()
	fmt.Printf("  Tx Hash:    %s\n", t.Hash)
	fmt.Printf("  Status:     %s\n", t.Status)
	if t.BlockNumber > 0 {
		fmt.Printf("  Block:      %d\n", t.BlockNumber)
	}
	if t.TimeHuman != "" {
		fmt.Printf("  Time:       %s\n", t.TimeHuman)
	}
	fmt.Printf("  From:       %s\n", t.From)
	fmt.Printf("  To:         %s\n", t.To)
	fmt.Printf("  Value:      %s\n", t.ValueFormatted)
	fmt.Printf("  Gas Price:  %s\n", t.GasPrice)
	fmt.Printf("  Gas Used:   %d / %d\n", t.GasUsed, t.GasLimit)
	fmt.Printf("  Tx Fee:     %s\n", t.TxFee)
	fmt.Printf("  Nonce:      %d\n", t.Nonce)
	if t.MethodName != "" {
		fmt.Printf("  Method:     %s\n", t.MethodName)
	} else if t.MethodID != "" {
		fmt.Printf("  Method ID:  %s\n", t.MethodID)
	}
	fmt.Printf("  Logs:       %d events\n", t.LogCount)
	fmt.Println()
}

func printBlockHuman(b domain.BlockInfo) {
	fmt.Println()
	fmt.Printf("  Block:       #%d\n", b.Number)
	fmt.Printf("  Hash:        %s\n", b.Hash)
	fmt.Printf("  Parent:      %s\n", b.ParentHash)
	fmt.Printf("  Time:        %s\n", b.TimeHuman)
	fmt.Printf("  Miner:       %s\n", b.Miner)
	fmt.Printf("  Gas Used:    %d / %d\n", b.GasUsed, b.GasLimit)
	if b.BaseFee != "" {
		fmt.Printf("  Base Fee:    %s\n", b.BaseFee)
	}
	fmt.Printf("  Txs:         %d transactions\n", b.TxCount)
	fmt.Println()
}

func printChainInfoHuman(c domain.ChainStatus) {
	fmt.Println()
	fmt.Printf("  Chain:        %s\n", c.Name)
	fmt.Printf("  Chain ID:     %d\n", c.ChainID)
	fmt.Printf("  RPC URL:      %s\n", c.RPCURL)
	fmt.Printf("  Currency:     %s\n", c.Currency)
	fmt.Printf("  Latest Block: %d\n", c.LatestBlock)
	fmt.Printf("  Gas Price:    %s\n", c.GasPrice)
	fmt.Println()
}

func printStakingHuman(s domain.StakingInfo) {
	fmt.Println()
	fmt.Printf("  Staking Summary for %s\n\n", shortenHash(s.Address))
	if len(s.Entries) == 0 {
		fmt.Println("  No staking positions found.")
	} else {
		fmt.Printf("  %-15s %-24s %s\n", "Validator", "Staked", "Claimable Rewards")
		fmt.Printf("  %s\n", strings.Repeat("-", 60))
		for _, e := range s.Entries {
			fmt.Printf("  %-15s %-24s %s\n", shortenHash(e.Validator), e.Staked, e.Rewards)
		}
		fmt.Println()
		fmt.Printf("  Total Staked:    %s\n", s.TotalStaked)
		fmt.Printf("  Total Rewards:   %s\n", s.TotalRewards)
	}
	fmt.Println()
}

func printValidatorsHuman(v domain.ValidatorsResult) {
	fmt.Println()
	fmt.Printf("  %s Validators (%d active)\n\n", v.Chain, v.Count)
	fmt.Printf("  %-4s %-15s %-10s %-22s %-12s %s\n", "#", "Address", "Status", "Delegated", "Commission", "Rewards")
	fmt.Printf("  %s\n", strings.Repeat("-", 80))
	for i, val := range v.Validators {
		addr := shortenHash(val.Address)
		fmt.Printf("  %-4d %-15s %-10s %-22s %-12s %s\n",
			i+1, addr, val.Status, val.TotalDelegated, val.CommissionRate, val.TotalRewards)
	}
	fmt.Println()
}

func printCallHuman(c domain.CallResult) {
	fmt.Println()
	if len(c.Values) > 0 {
		for _, v := range c.Values {
			fmt.Printf("  %s\n", v)
		}
	} else {
		fmt.Printf("  %s\n", c.Raw)
	}
	fmt.Println()
}

// shortenHash returns "0xabcd...ef12" format
func shortenHash(hash string) string {
	if len(hash) <= 14 {
		return hash
	}
	return hash[:6] + "..." + hash[len(hash)-4:]
}

// relativeTime returns a human-readable relative time string
func relativeTime(ts int64) string {
	if ts == 0 {
		return "pending"
	}
	diff := time.Now().Unix() - ts
	switch {
	case diff < 60:
		return fmt.Sprintf("%ds ago", diff)
	case diff < 3600:
		return fmt.Sprintf("%dm ago", diff/60)
	case diff < 86400:
		return fmt.Sprintf("%dh ago", diff/3600)
	default:
		return fmt.Sprintf("%dd ago", diff/86400)
	}
}
