package explorer

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/GrapeInTheTree/go-ethereum-butler/internal/domain"
)

// Client is an Etherscan-compatible block explorer API client (works with Chiliscan/Routescan)
type Client struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter <-chan time.Time
}

// NewClient creates a new explorer client with built-in rate limiting (2 req/sec)
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		rateLimiter: time.Tick(500 * time.Millisecond),
	}
}

// etherscanResponse is the generic Etherscan API response wrapper
type etherscanResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// rawTxListEntry maps a single transaction from the txlist endpoint
type rawTxListEntry struct {
	BlockNumber     string `json:"blockNumber"`
	TimeStamp       string `json:"timeStamp"`
	Hash            string `json:"hash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Gas             string `json:"gas"`
	GasPrice        string `json:"gasPrice"`
	Input           string `json:"input"`
	MethodID        string `json:"methodId"`
	FunctionName    string `json:"functionName"`
	TxReceiptStatus string `json:"txreceipt_status"`
	GasUsed         string `json:"gasUsed"`
	IsError         string `json:"isError"`
}

// rawTokenBalance maps a single token from the addresstokenbalance endpoint
type rawTokenBalance struct {
	TokenAddress string `json:"TokenAddress"`
	TokenName    string `json:"TokenName"`
	TokenSymbol  string `json:"TokenSymbol"`
	TokenDecimal string `json:"TokenDecimal"`
	Balance      string `json:"balance"`
}

// GetTxList retrieves normal transaction history for an address
func (c *Client) GetTxList(address string, page, offset int) ([]domain.TxSummary, error) {
	<-c.rateLimiter

	params := url.Values{
		"module":     {"account"},
		"action":     {"txlist"},
		"address":    {address},
		"startblock": {"0"},
		"endblock":   {"99999999"},
		"page":       {strconv.Itoa(page)},
		"offset":     {strconv.Itoa(offset)},
		"sort":       {"desc"},
	}

	var entries []rawTxListEntry
	if err := c.fetch(params, &entries); err != nil {
		return nil, fmt.Errorf("GetTxList: %w", err)
	}

	txs := make([]domain.TxSummary, 0, len(entries))
	for _, e := range entries {
		ts, _ := strconv.ParseInt(e.TimeStamp, 10, 64)
		txs = append(txs, domain.TxSummary{
			Hash:      e.Hash,
			Method:    methodName(e.FunctionName, e.MethodID),
			From:      e.From,
			To:        e.To,
			Value:     formatWei(e.Value),
			Timestamp: ts,
			TimeHuman: formatTime(ts),
			IsError:   e.IsError == "1",
			GasUsed:   e.GasUsed,
			TxFee:     calcTxFee(e.GasUsed, e.GasPrice),
		})
	}
	return txs, nil
}

// GetTokenBalances retrieves all ERC-20 token balances for an address
func (c *Client) GetTokenBalances(address string) ([]domain.TokenBalance, error) {
	<-c.rateLimiter

	params := url.Values{
		"module":  {"account"},
		"action":  {"addresstokenbalance"},
		"address": {address},
		"page":    {"1"},
		"offset":  {"100"},
	}

	var entries []rawTokenBalance
	if err := c.fetch(params, &entries); err != nil {
		return nil, fmt.Errorf("GetTokenBalances: %w", err)
	}

	balances := make([]domain.TokenBalance, 0, len(entries))
	for _, e := range entries {
		decimals, _ := strconv.Atoi(e.TokenDecimal)
		balances = append(balances, domain.TokenBalance{
			Symbol:          e.TokenSymbol,
			Name:            e.TokenName,
			ContractAddress: e.TokenAddress,
			Balance:         formatTokenAmount(e.Balance, decimals),
			Decimals:        decimals,
		})
	}
	return balances, nil
}

// GetTokenTxList retrieves ERC-20 token transfer history for an address
func (c *Client) GetTokenTxList(address string, page, offset int) ([]domain.TxSummary, error) {
	<-c.rateLimiter

	params := url.Values{
		"module":     {"account"},
		"action":     {"tokentx"},
		"address":    {address},
		"startblock": {"0"},
		"endblock":   {"99999999"},
		"page":       {strconv.Itoa(page)},
		"offset":     {strconv.Itoa(offset)},
		"sort":       {"desc"},
	}

	var entries []rawTxListEntry
	if err := c.fetch(params, &entries); err != nil {
		return nil, fmt.Errorf("GetTokenTxList: %w", err)
	}

	txs := make([]domain.TxSummary, 0, len(entries))
	for _, e := range entries {
		ts, _ := strconv.ParseInt(e.TimeStamp, 10, 64)
		txs = append(txs, domain.TxSummary{
			Hash:      e.Hash,
			Method:    methodName(e.FunctionName, e.MethodID),
			From:      e.From,
			To:        e.To,
			Value:     e.Value,
			Timestamp: ts,
			TimeHuman: formatTime(ts),
			IsError:   e.IsError == "1",
			GasUsed:   e.GasUsed,
			TxFee:     calcTxFee(e.GasUsed, e.GasPrice),
		})
	}
	return txs, nil
}

// fetch executes an API call and unmarshals the result into target
func (c *Client) fetch(params url.Values, target any) error {
	reqURL := c.baseURL + "?" + params.Encode()
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp etherscanResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Status != "1" {
		// Status "0" with "No transactions found" is not an error — just empty results
		if apiResp.Message == "No transactions found" || apiResp.Message == "No records found" {
			return nil
		}
		return fmt.Errorf("API error: %s", apiResp.Message)
	}

	if err := json.Unmarshal(apiResp.Result, target); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}
	return nil
}

// formatWei converts a wei string to a human-readable ether value (18 decimals)
func formatWei(weiStr string) string {
	return formatTokenAmount(weiStr, 18)
}

// formatTokenAmount converts a raw token amount string to human-readable format
func formatTokenAmount(rawAmount string, decimals int) string {
	bal := new(big.Int)
	bal.SetString(rawAmount, 10)
	if bal.Sign() == 0 {
		return "0"
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	fBal := new(big.Float).SetInt(bal)
	fDiv := new(big.Float).SetInt(divisor)
	result := new(big.Float).Quo(fBal, fDiv)
	return result.Text('f', 6)
}

// calcTxFee computes gasUsed * gasPrice in ether
func calcTxFee(gasUsedStr, gasPriceStr string) string {
	gasUsed := new(big.Int)
	gasUsed.SetString(gasUsedStr, 10)
	gasPrice := new(big.Int)
	gasPrice.SetString(gasPriceStr, 10)
	fee := new(big.Int).Mul(gasUsed, gasPrice)
	return formatWei(fee.String())
}

// methodName returns a readable method name
func methodName(funcName, methodID string) string {
	if funcName != "" {
		return funcName
	}
	if methodID != "" && methodID != "0x" {
		return methodID
	}
	return "Transfer"
}

// formatTime converts a unix timestamp to a human-readable string
func formatTime(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format("2006-01-02 15:04:05 UTC")
}
