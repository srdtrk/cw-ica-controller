package types

import "encoding/base64"

type ContractCosmosMsg struct {
	Stargate     *StargateCosmosMsg     `json:"stargate,omitempty"`
	Bank         *BankCosmosMsg         `json:"bank,omitempty"`
	IBC          *IbcCosmosMsg          `json:"ibc,omitempty"`
	Staking      *StakingCosmosMsg      `json:"staking,omitempty"`
	Distribution *DistributionCosmosMsg `json:"distribution,omitempty"`
	Gov          *GovCosmosMsg          `json:"gov,omitempty"`
	Wasm         *WasmCosmosMsg         `json:"wasm,omitempty"`
}

type StargateCosmosMsg struct {
	// Proto Any type URL
	TypeUrl string `json:"type_url"`
	// Base64 encoded bytes
	Value string `json:"value"`
}

type BankCosmosMsg struct {
	Send *BankSendCosmosMsg `json:"send,omitempty"`
}

type IbcCosmosMsg struct {
	Transfer *IbcTransferCosmosMsg `json:"transfer,omitempty"`
}

type GovCosmosMsg struct {
	Vote         *GovVoteCosmosMsg         `json:"vote,omitempty"`
	VoteWeighted *GovVoteWeightedCosmosMsg `json:"vote_weighted,omitempty"`
}

type StakingCosmosMsg struct {
	Delegate   *StakingDelegateCosmosMsg   `json:"delegate,omitempty"`
	Undelegate *StakingUndelegateCosmosMsg `json:"undelegate,omitempty"`
	Redelegate *StakingRedelegateCosmosMsg `json:"redelegate,omitempty"`
}

type DistributionCosmosMsg struct {
	SetWithdrawAddress      *DistributionSetWithdrawAddressCosmosMsg      `json:"set_withdraw_address,omitempty"`
	WithdrawDelegatorReward *DistributionWithdrawDelegatorRewardCosmosMsg `json:"withdraw_delegator_reward,omitempty"`
	FundCommunityPool       *DistributionFundCommunityPoolCosmosMsg       `json:"fund_community_pool,omitempty"`
}

type WasmCosmosMsg struct {
	Execute      *WasmExecuteCosmosMsg      `json:"execute,omitempty"`
	Instantiate  *WasmInstantiateCosmosMsg  `json:"instantiate,omitempty"`
	Instantiate2 *WasmInstantiate2CosmosMsg `json:"instantiate2,omitempty"`
	Migrate      *WasmMigrateCosmosMsg      `json:"migrate,omitempty"`
	UpdateAdmin  *WasmUpdateAdminCosmosMsg  `json:"update_admin,omitempty"`
	ClearAdmin   *WasmClearAdminCosmosMsg   `json:"clear_admin,omitempty"`
}

type WasmExecuteCosmosMsg struct {
	ContractAddr string `json:"contract_addr"`
	// base64 encoded bytes
	Msg   string `json:"msg"`
	Funds []Coin `json:"funds"`
}

type WasmInstantiateCosmosMsg struct {
	Admin  string `json:"admin"`
	CodeID uint64 `json:"code_id"`
	// base64 encoded bytes
	Msg   string `json:"msg"`
	Funds []Coin `json:"funds"`
	Label string `json:"label"`
}

type WasmInstantiate2CosmosMsg struct {
	Admin  string `json:"admin"`
	CodeID uint64 `json:"code_id"`
	// base64 encoded bytes
	Msg   string `json:"msg"`
	Funds []Coin `json:"funds"`
	Label string `json:"label"`
	// base64 encoded bytes
	Salt string `json:"salt"`
}

type WasmMigrateCosmosMsg struct {
	ContractAddr string `json:"contract_addr"`
	NewCodeID    uint64 `json:"new_code_id"`
	// base64 encoded bytes
	Msg string `json:"msg"`
}

type WasmUpdateAdminCosmosMsg struct {
	ContractAddr string `json:"contract_addr"`
	Admin        string `json:"admin"`
}

type WasmClearAdminCosmosMsg struct {
	ContractAddr string `json:"contract_addr"`
}

type DistributionSetWithdrawAddressCosmosMsg struct {
	Address string `json:"address"`
}

type DistributionWithdrawDelegatorRewardCosmosMsg struct {
	Validator string `json:"validator"`
}

type DistributionFundCommunityPoolCosmosMsg struct {
	Amount []Coin `json:"amount"`
}

type StakingDelegateCosmosMsg struct {
	Validator string `json:"validator"`
	Amount    Coin   `json:"amount"`
}

type StakingUndelegateCosmosMsg struct {
	Validator string `json:"validator"`
	Amount    Coin   `json:"amount"`
}

type StakingRedelegateCosmosMsg struct {
	SrcValidator string `json:"src_validator"`
	DstValidator string `json:"dst_validator"`
	Amount       Coin   `json:"amount"`
}

type BankSendCosmosMsg struct {
	ToAddress string `json:"to_address"`
	Amount    []Coin `json:"amount"`
}

type IbcTransferCosmosMsg struct {
	ChannelID string `json:"channel_id"`
	ToAddress string `json:"to_address"`
	Amount    Coin   `json:"amount"`
	// Timeout string `json:"timeout"`
}

type GovVoteCosmosMsg struct {
	ProposalID uint64 `json:"proposal_id"`
	Vote       string `json:"vote"`
}

type GovVoteWeightedCosmosMsg struct {
	ProposalID uint64                  `json:"proposal_id"`
	Options    []GovVoteWeightedOption `json:"options"`
}

type GovVoteWeightedOption struct {
	Option string `json:"option"`
	Weight string `json:"weight"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// NewCosmosMsgWithStargate creates a new CosmosMsg with a Stargate message
func NewCosmosMsgWithStargate(typeUrl string, value []byte) *ContractCosmosMsg {
	return &ContractCosmosMsg{
		Stargate: &StargateCosmosMsg{
			TypeUrl: typeUrl,
			Value:   base64.StdEncoding.EncodeToString(value),
		},
	}
}

// NewCosmosMsgWithBankSend creates a new CosmosMsg with a Bank message
func NewCosmosMsgWithBankSend(toAddress string, amount []Coin) *ContractCosmosMsg {
	return &ContractCosmosMsg{
		Bank: &BankCosmosMsg{
			Send: &BankSendCosmosMsg{
				ToAddress: toAddress,
				Amount:    amount,
			},
		},
	}
}

// NewCosmosMsgWithIbcTransfer creates a new CosmosMsg with an IBC transfer message
func NewCosmosMsgWithIbcTransfer(channelID string, toAddress string, amount Coin) *ContractCosmosMsg {
	return &ContractCosmosMsg{
		IBC: &IbcCosmosMsg{
			Transfer: &IbcTransferCosmosMsg{
				ChannelID: channelID,
				ToAddress: toAddress,
				Amount:    amount,
			},
		},
	}
}
