package domain

// AddressInfo is the response for `butler address <addr>`
type AddressInfo struct {
	Address       string         `json:"address"`
	Chain         string         `json:"chain"`
	ChainID       int64          `json:"chain_id"`
	NativeBalance string         `json:"native_balance"`
	NativeSymbol  string         `json:"native_symbol"`
	Nonce         uint64         `json:"nonce"`
	IsContract    bool           `json:"is_contract"`
	TokenBalances []TokenBalance `json:"token_balances,omitempty"`
	RecentTxs     []TxSummary    `json:"recent_txs,omitempty"`
}

// TokenBalance represents a single token holding
type TokenBalance struct {
	Symbol          string `json:"symbol"`
	Name            string `json:"name"`
	ContractAddress string `json:"contract_address"`
	Balance         string `json:"balance"`
	Decimals        int    `json:"decimals"`
}

// TxSummary represents a transaction in a list
type TxSummary struct {
	Hash      string `json:"hash"`
	Method    string `json:"method,omitempty"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	TimeHuman string `json:"time_human"`
	IsError   bool   `json:"is_error"`
	GasUsed   string `json:"gas_used"`
	TxFee     string `json:"tx_fee,omitempty"`
}

// TxDetail is the response for `butler tx <hash>`
type TxDetail struct {
	Hash           string `json:"hash"`
	Status         string `json:"status"` // "success", "failed", "pending"
	BlockNumber    uint64 `json:"block_number"`
	Timestamp      int64  `json:"timestamp,omitempty"`
	TimeHuman      string `json:"time_human,omitempty"`
	From           string `json:"from"`
	To             string `json:"to"`
	Value          string `json:"value"`
	ValueFormatted string `json:"value_formatted"`
	GasPrice       string `json:"gas_price"`
	GasUsed        uint64 `json:"gas_used"`
	GasLimit       uint64 `json:"gas_limit"`
	TxFee          string `json:"tx_fee"`
	Nonce          uint64 `json:"nonce"`
	InputData      string `json:"input_data"`
	MethodID       string `json:"method_id,omitempty"`
	LogCount       int    `json:"log_count"`
}

// BlockInfo is the response for `butler block [number]`
type BlockInfo struct {
	Number     uint64 `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parent_hash"`
	Timestamp  int64  `json:"timestamp"`
	TimeHuman  string `json:"time_human"`
	GasUsed    uint64 `json:"gas_used"`
	GasLimit   uint64 `json:"gas_limit"`
	TxCount    int    `json:"tx_count"`
	Miner      string `json:"miner"`
	BaseFee    string `json:"base_fee,omitempty"`
}

// ChainStatus is the response for `butler chain-info`
type ChainStatus struct {
	Name        string `json:"name"`
	ChainID     int64  `json:"chain_id"`
	RPCURL      string `json:"rpc_url"`
	LatestBlock uint64 `json:"latest_block"`
	GasPrice    string `json:"gas_price"`
	Currency    string `json:"currency_symbol"`
}

// ValidatorInfo represents a single validator's status
type ValidatorInfo struct {
	Address        string `json:"address"`
	Owner          string `json:"owner"`
	Status         string `json:"status"`
	TotalDelegated string `json:"total_delegated"`
	SlashCount     uint32 `json:"slash_count"`
	CommissionRate string `json:"commission_rate"`
	TotalRewards   string `json:"total_rewards"`
}

// ValidatorsResult is the response for `butler validators`
type ValidatorsResult struct {
	Chain      string          `json:"chain"`
	ChainID    int64           `json:"chain_id"`
	Count      int             `json:"count"`
	Validators []ValidatorInfo `json:"validators"`
}

// CallResult is the response for `butler call <contract> <sig> [args...]`
type CallResult struct {
	Contract string   `json:"contract"`
	Method   string   `json:"method"`
	Values   []string `json:"values,omitempty"`
	Raw      string   `json:"raw"`
}
