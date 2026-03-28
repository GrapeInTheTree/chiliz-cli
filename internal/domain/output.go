package domain

// AddressInfo is the response for `chiliz address <addr>`
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
	InternalTxs   []TxSummary    `json:"internal_txs,omitempty"`
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

// TxDetail is the response for `chiliz tx <hash>`
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
	MethodName     string      `json:"method_name,omitempty"`
	LogCount       int         `json:"log_count"`
	InternalTxs    []TxSummary `json:"internal_txs,omitempty"`
}

// BlockInfo is the response for `chiliz block [number]`
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

// ChainStatus is the response for `chiliz chain-info`
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
	VotingPower    string `json:"voting_power"`
	APY            string `json:"apy"`
	CommissionRate string `json:"commission_rate"`
	TotalRewards   string `json:"total_rewards"`
	SlashCount     uint32 `json:"slash_count"`
}

// ValidatorsResult is the response for `chiliz validators`
type ValidatorsResult struct {
	Chain      string          `json:"chain"`
	ChainID    int64           `json:"chain_id"`
	Count      int             `json:"count"`
	Validators []ValidatorInfo `json:"validators"`
}

// ContractInfo is the response for `chiliz contract <address>`
type ContractInfo struct {
	Address        string `json:"address"`
	Name           string `json:"name"`
	Compiler       string `json:"compiler,omitempty"`
	EVMVersion     string `json:"evm_version,omitempty"`
	Optimized      bool   `json:"optimized"`
	Runs           string `json:"runs,omitempty"`
	License        string `json:"license,omitempty"`
	IsProxy        bool   `json:"is_proxy"`
	Implementation string `json:"implementation,omitempty"`
	Verified       bool   `json:"verified"`
	Deployer       string `json:"deployer,omitempty"`
	DeployTx       string `json:"deploy_tx,omitempty"`
}

// HolderEntry represents a single token holder
type HolderEntry struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

// HoldersResult is the response for `chiliz holders <token>`
type HoldersResult struct {
	Token      string        `json:"token"`
	TotalCount string        `json:"total_count"`
	Holders    []HolderEntry `json:"holders"`
}

// LogEntry represents a single event log
type LogEntry struct {
	Address     string   `json:"address"`
	BlockNumber uint64   `json:"block_number"`
	TxHash      string   `json:"tx_hash"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	EventName   string   `json:"event_name,omitempty"`
}

// LogsResult is the response for `chiliz logs`
type LogsResult struct {
	Address string     `json:"address"`
	Event   string     `json:"event,omitempty"`
	Count   int        `json:"count"`
	Logs    []LogEntry `json:"logs"`
}

// TokenDetail is the response for `chiliz token <address>`
type TokenDetail struct {
	ContractAddress string `json:"contract_address"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	TokenType       string `json:"token_type"`
	Decimals        int    `json:"decimals"`
	TotalSupply     string `json:"total_supply"`
	Verified        bool   `json:"verified"`
	PriceUSD        string `json:"price_usd,omitempty"`
	Website         string `json:"website,omitempty"`
	Twitter         string `json:"twitter,omitempty"`
	Telegram        string `json:"telegram,omitempty"`
}

// StakingEntry represents staking for a single validator
type StakingEntry struct {
	Validator  string `json:"validator"`
	Staked     string `json:"staked"`
	Rewards    string `json:"rewards"`
}

// StakingInfo is the response for `chiliz staking <address>`
type StakingInfo struct {
	Address      string         `json:"address"`
	Chain        string         `json:"chain"`
	TotalStaked  string         `json:"total_staked"`
	TotalRewards string         `json:"total_rewards"`
	Entries      []StakingEntry `json:"entries,omitempty"`
}

// CallResult is the response for `chiliz call <contract> <sig> [args...]`
type CallResult struct {
	Contract string   `json:"contract"`
	Method   string   `json:"method"`
	Values   []string `json:"values,omitempty"`
	Raw      string   `json:"raw"`
}
