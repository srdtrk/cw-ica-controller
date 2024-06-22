/* Code generated by github.com/srdtrk/go-codegen, DO NOT EDIT. */
package callbackcounter

type InstantiateMsg struct{}

// This is the execute message of the contract.
type ExecuteMsg struct {
	// The callback message from `cw-ica-controller`. The handler for this variant should verify that this message comes from an expected legitimate source.
	ReceiveIcaCallback *ExecuteMsg_ReceiveIcaCallback `json:"receive_ica_callback,omitempty"`
}

type QueryMsg struct {
	// GetCallbackCounter returns the callback counter.
	GetCallbackCounter *QueryMsg_GetCallbackCounter `json:"get_callback_counter,omitempty"`
}

type IbcEndpoint struct {
	ChannelId string `json:"channel_id"`
	PortId string `json:"port_id"`
}

// `Data` is the response to an ibc packet. It either contains a result or an error.
type Data struct {
	// Result is the result of a successful transaction.
	Result *Data_Result `json:"result,omitempty"`
	// Error is the error message of a failed transaction. It is a string of the error message (not base64 encoded).
	Error *Data_Error `json:"error,omitempty"`
}

// The response type for the [`cosmwasm_std::BankQuery`] queries.
type BankQueryResponse struct {
	// Response for the [`cosmwasm_std::BankQuery::Supply`] query.
	Supply *BankQueryResponse_Supply `json:"supply,omitempty"`
	// Response for the [`cosmwasm_std::BankQuery::Balance`] query.
	Balance *BankQueryResponse_Balance `json:"balance,omitempty"`
	// Response for the [`cosmwasm_std::BankQuery::AllBalances`] query.
	AllBalances *BankQueryResponse_AllBalances `json:"all_balances,omitempty"`
	// Response for the [`cosmwasm_std::BankQuery::DenomMetadata`] query.
	DenomMetadata *BankQueryResponse_DenomMetadata `json:"denom_metadata,omitempty"`
	// Response for the [`cosmwasm_std::BankQuery::AllDenomMetadata`] query.
	AllDenomMetadata *BankQueryResponse_AllDenomMetadata `json:"all_denom_metadata,omitempty"`
}

/*
A fixed-point decimal value with 18 fractional digits, i.e. Decimal(1_000_000_000_000_000_000) == 1.0

The greatest possible value that can be represented is 340282366920938463463.374607431768211455 (which is (2^128 - 1) / 10^18)
*/
type Decimal string

/*
Binary is a wrapper around Vec<u8> to add base64 de/serialization with serde. It also adds some helper methods to help encode inline.

This is only needed as serde-json-{core,wasm} has a horrible encoding for Vec<u8>. See also <https://github.com/CosmWasm/cosmwasm/blob/main/docs/MESSAGE_TYPES.md>.
*/
type Binary string

// IbcOrder defines if a channel is ORDERED or UNORDERED Values come from https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/ibc/core/channel/v1/channel.proto#L69-L80 Naming comes from the protobuf files and go translations.
type IbcOrder string

const (
	IbcOrder_OrderUnordered IbcOrder = "ORDER_UNORDERED"
	IbcOrder_OrderOrdered   IbcOrder = "ORDER_ORDERED"
)

// Response for the [`cosmwasm_std::StakingQuery::AllDelegations`] query over ICA.
type IcaAllDelegationsResponse struct {
	// The delegations.
	Delegations []Delegation `json:"delegations"`
}

// Replicates the cosmos-sdk bank module DenomUnit type
type DenomUnit struct {
	Aliases []string `json:"aliases"`
	Denom string `json:"denom"`
	Exponent int `json:"exponent"`
}

type IbcPacket struct {
	// identifies the channel and port on the sending chain.
	Src IbcEndpoint `json:"src"`
	Timeout IbcTimeout `json:"timeout"`
	// The raw data sent from the other side in the packet
	Data Binary `json:"data"`
	// identifies the channel and port on the receiving chain.
	Dest IbcEndpoint `json:"dest"`
	// The sequence number of the packet on the given channel
	Sequence int `json:"sequence"`
}

// Response for the [`cosmwasm_std::StakingQuery::Delegation`] query over ICA.
type IcaDelegationResponse struct {
	// The delegation response if it exists.
	Delegation *Delegation `json:"delegation,omitempty"`
}

type Coin struct {
	Amount Uint128 `json:"amount"`
	Denom string `json:"denom"`
}

/*
A point in time in nanosecond precision.

This type can represent times from 1970-01-01T00:00:00Z to 2554-07-21T23:34:33Z.

## Examples

``` # use cosmwasm_std::Timestamp; let ts = Timestamp::from_nanos(1_000_000_202); assert_eq!(ts.nanos(), 1_000_000_202); assert_eq!(ts.seconds(), 1); assert_eq!(ts.subsec_nanos(), 202);

let ts = ts.plus_seconds(2); assert_eq!(ts.nanos(), 3_000_000_202); assert_eq!(ts.seconds(), 3); assert_eq!(ts.subsec_nanos(), 202); ```
*/
type Timestamp Uint64

type SupplyResponse struct {
	// Always returns a Coin with the requested denom. This will be of zero amount if the denom does not exist.
	Amount Coin `json:"amount"`
}

// `TxEncoding` is the encoding of the transactions sent to the ICA host.
type TxEncoding string

const (
	// `Protobuf` is the protobuf serialization of the CosmosSDK's Any.
	TxEncoding_Proto3 TxEncoding = "proto3"
	// `Proto3Json` is the json serialization of the CosmosSDK's Any.
	TxEncoding_Proto3Json TxEncoding = "proto3json"
)

type AllBalanceResponse struct {
	// Returns all non-zero coins held by this account.
	Amount []Coin `json:"amount"`
}

// The response type for the [`cosmwasm_std::StakingQuery`] queries.
type StakingQueryResponse struct {
	// Response for the [`cosmwasm_std::StakingQuery::BondedDenom`] query.
	BondedDenom *StakingQueryResponse_BondedDenom `json:"bonded_denom,omitempty"`
	// Response for the [`cosmwasm_std::StakingQuery::AllDelegations`] query.
	AllDelegations *StakingQueryResponse_AllDelegations `json:"all_delegations,omitempty"`
	// Response for the [`cosmwasm_std::StakingQuery::Delegation`] query.
	Delegation *StakingQueryResponse_Delegation `json:"delegation,omitempty"`
	// Response for the [`cosmwasm_std::StakingQuery::AllValidators`] query.
	AllValidators *StakingQueryResponse_AllValidators `json:"all_validators,omitempty"`
	// Response for the [`cosmwasm_std::StakingQuery::Validator`] query.
	Validator *StakingQueryResponse_Validator `json:"validator,omitempty"`
}

// The response for a successful ICA query.
type IcaQueryResponse struct {
	// Response for a [`cosmwasm_std::BankQuery`].
	Bank *IcaQueryResponse_Bank `json:"bank,omitempty"`
	// Response for a [`cosmwasm_std::QueryRequest::Stargate`]. Protobuf encoded bytes stored as [`cosmwasm_std::Binary`].
	Stargate *IcaQueryResponse_Stargate `json:"stargate,omitempty"`
	// Response for a [`cosmwasm_std::StakingQuery`].
	Staking *IcaQueryResponse_Staking `json:"staking,omitempty"`
}

// `IcaControllerCallbackMsg` is the type of message that this contract can send to other contracts.
type IcaControllerCallbackMsg struct {
	// `OnAcknowledgementPacketCallback` is the callback that this contract makes to other contracts when it receives an acknowledgement packet.
	OnAcknowledgementPacketCallback *IcaControllerCallbackMsg_OnAcknowledgementPacketCallback `json:"on_acknowledgement_packet_callback,omitempty"`
	// `OnTimeoutPacketCallback` is the callback that this contract makes to other contracts when it receives a timeout packet.
	OnTimeoutPacketCallback *IcaControllerCallbackMsg_OnTimeoutPacketCallback `json:"on_timeout_packet_callback,omitempty"`
	// `OnChannelOpenAckCallback` is the callback that this contract makes to other contracts when it receives a channel open acknowledgement.
	OnChannelOpenAckCallback *IcaControllerCallbackMsg_OnChannelOpenAckCallback `json:"on_channel_open_ack_callback,omitempty"`
}

// The data format returned from StakingRequest::AllValidators query
type AllValidatorsResponse struct {
	Validators []Validator `json:"validators"`
}

// CallbackCounter tracks the number of callbacks in store.
type CallbackCounter struct {
	// The erroneous callbacks.
	Error []IcaControllerCallbackMsg `json:"error"`
	// The successful callbacks.
	Success []IcaControllerCallbackMsg `json:"success"`
	// The timeout callbacks. The channel is closed after a timeout if the channel is ordered due to the semantics of ordered channels.
	Timeout []IcaControllerCallbackMsg `json:"timeout"`
}

// BondedDenomResponse is data format returned from StakingRequest::BondedDenom query
type BondedDenomResponse struct {
	Denom string `json:"denom"`
}

/*
A human readable address.

In Cosmos, this is typically bech32 encoded. But for multi-chain smart contracts no assumptions should be made other than being UTF-8 encoded and of reasonable length.

This type represents a validated address. It can be created in the following ways 1. Use `Addr::unchecked(input)` 2. Use `let checked: Addr = deps.api.addr_validate(input)?` 3. Use `let checked: Addr = deps.api.addr_humanize(canonical_addr)?` 4. Deserialize from JSON. This must only be done from JSON that was validated before such as a contract's state. `Addr` must not be used in messages sent by the user because this would result in unvalidated instances.

This type is immutable. If you really need to mutate it (Really? Are you sure?), create a mutable copy using `let mut mutable = Addr::to_string()` and operate on that `String` instance.
*/
type Addr string

/*
A thin wrapper around u128 that is using strings for JSON encoding/decoding, such that the full u128 range can be used for clients that convert JSON numbers to floats, like JavaScript and jq.

# Examples

Use `from` to create instances of this and `u128` to get the value out:

``` # use cosmwasm_std::Uint128; let a = Uint128::from(123u128); assert_eq!(a.u128(), 123);

let b = Uint128::from(42u64); assert_eq!(b.u128(), 42);

let c = Uint128::from(70u32); assert_eq!(c.u128(), 70); ```
*/
type Uint128 string

// In IBC each package must set at least one type of timeout: the timestamp or the block height. Using this rather complex enum instead of two timeout fields we ensure that at least one timeout is set.
type IbcTimeout struct {
	Block *IbcTimeoutBlock `json:"block,omitempty"`
	Timestamp *Timestamp `json:"timestamp,omitempty"`
}

type DenomMetadataResponse struct {
	// The metadata for the queried denom.
	Metadata DenomMetadata `json:"metadata"`
}

type QueryMsg_GetCallbackCounter struct{}

// Instances are created in the querier.
type Validator struct {
	/*
	   The operator address of the validator (e.g. cosmosvaloper1...). See https://github.com/cosmos/cosmos-sdk/blob/v0.47.4/proto/cosmos/staking/v1beta1/staking.proto#L95-L96 for more information.

	   This uses `String` instead of `Addr` since the bech32 address prefix is different from the ones that regular user accounts use.
	*/
	Address string `json:"address"`
	Commission Decimal `json:"commission"`
	// The maximum daily increase of the commission
	MaxChangeRate Decimal `json:"max_change_rate"`
	MaxCommission Decimal `json:"max_commission"`
}

// The data format returned from StakingRequest::Validator query
type ValidatorResponse struct {
	Validator *Validator `json:"validator,omitempty"`
}

// The result of an ICA query packet.
type IcaQueryResult struct {
	// The query was successful and the responses are included.
	Success *IcaQueryResult_Success `json:"success,omitempty"`
	// The query failed with an error message. The error string often does not contain useful information for the end user.
	Error *IcaQueryResult_Error `json:"error,omitempty"`
}

// Delegation is the detailed information about a delegation.
type Delegation struct {
	// Delegation amount.
	Amount Coin `json:"amount"`
	// The delegator address.
	Delegator string `json:"delegator"`
	// A validator address (e.g. cosmosvaloper1...)
	Validator string `json:"validator"`
}

// IBCTimeoutHeight Height is a monotonically increasing data type that can be compared against another Height for the purposes of updating and freezing clients. Ordering is (revision_number, timeout_height)
type IbcTimeoutBlock struct {
	// block height after which the packet times out. the height within the given revision
	Height int `json:"height"`
	// the version that the client is currently on (e.g. after resetting the chain this could increment 1 as height drops to 0)
	Revision int `json:"revision"`
}

type AllDenomMetadataResponse struct {
	// Always returns metadata for all token denoms on the base chain.
	Metadata []DenomMetadata `json:"metadata"`
	NextKey *Binary `json:"next_key,omitempty"`
}
type ExecuteMsg_ReceiveIcaCallback IcaControllerCallbackMsg

type BalanceResponse struct {
	// Always returns a Coin with the requested denom. This may be of 0 amount if no such funds.
	Amount Coin `json:"amount"`
}

// IbcChannel defines all information on a channel. This is generally used in the hand-shake process, but can be queried directly.
type IbcChannel struct {
	Order IbcOrder `json:"order"`
	// Note: in ibcv3 this may be "", in the IbcOpenChannel handshake messages
	Version string `json:"version"`
	// The connection upon which this channel was created. If this is a multi-hop channel, we only expose the first hop.
	ConnectionId string `json:"connection_id"`
	CounterpartyEndpoint IbcEndpoint `json:"counterparty_endpoint"`
	Endpoint IbcEndpoint `json:"endpoint"`
}

// Replicates the cosmos-sdk bank module Metadata type
type DenomMetadata struct {
	Base string `json:"base"`
	DenomUnits []DenomUnit `json:"denom_units"`
	Description string `json:"description"`
	Display string `json:"display"`
	Name string `json:"name"`
	Symbol string `json:"symbol"`
	Uri string `json:"uri"`
	UriHash string `json:"uri_hash"`
}

/*
A thin wrapper around u64 that is using strings for JSON encoding/decoding, such that the full u64 range can be used for clients that convert JSON numbers to floats, like JavaScript and jq.

# Examples

Use `from` to create instances of this and `u64` to get the value out:

``` # use cosmwasm_std::Uint64; let a = Uint64::from(42u64); assert_eq!(a.u64(), 42);

let b = Uint64::from(70u32); assert_eq!(b.u64(), 70); ```
*/
type Uint64 string
type BankQueryResponse_DenomMetadata DenomMetadataResponse

type Data_Error string
type Data_Result Binary

type IcaControllerCallbackMsg_OnAcknowledgementPacketCallback struct {
	// The responses to the queries.
	QueryResult *IcaQueryResult `json:"query_result,omitempty"`
	// The relayer that submitted acknowledgement packet
	Relayer Addr `json:"relayer"`
	// The deserialized ICA acknowledgement data
	IcaAcknowledgement Data `json:"ica_acknowledgement"`
	// The original packet that was sent
	OriginalPacket IbcPacket `json:"original_packet"`
}

type IcaControllerCallbackMsg_OnChannelOpenAckCallback struct {
	// The channel that was opened.
	Channel IbcChannel `json:"channel"`
	// The address of the interchain account that was created.
	IcaAddress string `json:"ica_address"`
	// The tx encoding this ICA channel uses.
	TxEncoding TxEncoding `json:"tx_encoding"`
}
type StakingQueryResponse_BondedDenom BondedDenomResponse
type StakingQueryResponse_AllDelegations IcaAllDelegationsResponse
type IcaQueryResponse_Bank BankQueryResponse

type IcaQueryResult_Error string
type BankQueryResponse_Supply SupplyResponse
type IcaQueryResponse_Staking StakingQueryResponse

type IcaControllerCallbackMsg_OnTimeoutPacketCallback struct {
	// The relayer that submitted acknowledgement packet
	Relayer Addr `json:"relayer"`
	// The original packet that was sent
	OriginalPacket IbcPacket `json:"original_packet"`
}
type BankQueryResponse_AllBalances AllBalanceResponse
type BankQueryResponse_AllDenomMetadata AllDenomMetadataResponse
type StakingQueryResponse_Delegation IcaDelegationResponse

type IcaQueryResponse_Stargate struct {
	// The response bytes.
	Data Binary `json:"data"`
	// The query grpc method
	Path string `json:"path"`
}
type BankQueryResponse_Balance BalanceResponse
type StakingQueryResponse_AllValidators AllValidatorsResponse
type StakingQueryResponse_Validator ValidatorResponse

type IcaQueryResult_Success struct {
	// The height of the block at which the queries were executed on the counterparty chain.
	Height int `json:"height"`
	// The responses to the queries.
	Responses []IcaQueryResponse `json:"responses"`
}
