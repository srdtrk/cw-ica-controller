# CosmWasm ICA Controller Contract

![cw-ica-controller](./cw-ica-controller.svg)

This is a CosmWasm smart contract that communicates with the golang ica/host module on the host chain to create and manage one interchain account. This contract can also execute callbacks based on the result of the interchain account transaction. Because this is a CosmWasm implementation of the entire ICA controller, the chain that this contract is deployed on need **not** have the ICA module enabled. This contract can be deployed on any chain that supports IBC CosmWasm smart contracts.

This contract was originally written to test the json encoding/decoding [feature being added to interchain accounts txs](https://github.com/cosmos/ibc-go/pull/3796). Previously, the only way to encode ica txs was to use Protobuf (`proto3`) encoding. Proto3Json encoding/decoding feature is supported in ibc-go v7.3+. This contract now supports both proto3json and protobuf encoding/decoding. In current mainnets (ibc-go v7.2 and below), only protobuf encoding/decoding is supported.

## Usage

The following is a brief overview of the contract's functionality. You can also see the various ways this contract can be used in the end to end tests in the `e2e` directory.

### Create an interchain account

To create an interchain account, the relayer must start the channel handshake on the contract's chain. See end to end tests for an example of how to do this. Unfortunately, this is not possible to do in the contract itself. Also, you cannot initialize the channel handshake with an empty string as the version, this is due to a limitation of the IBCModule interface provided by ibc-go, see issue [#3942](https://github.com/cosmos/ibc-go/issues/3942). (The contract can now be initialized with an empty version string if the chain supports stargate queries, but this is not the case for the end to end tests.) The version string we are using for tests is: `{"version":"ics27-1","controller_connection_id":"connection-0","host_connection_id":"connection-0","address":"","encoding":"proto3json","tx_type":"sdk_multi_msg"}` (encoding is replaced with `"proto3"` to test protobuf encoding/decoding). You can see all this in the [end to end tests](./e2e/).

### Execute an interchain account transaction

In this contract, the `execute` message is used to commit a packet to be sent to the host chain. This contract has two ways of executing an interchain transaction:

1. `SendCustomIcaMessages`: This message requires the sender to give base64 encoded messages that will be sent to the host chain. The host chain will decode the messages and execute them. The result of the execution is sent back to this contract, and a callback is executed based on the result.

If the channel is using `proto3json` encoding, then the format that json messages have to take are defined by the cosmos-sdk's json codec. The following is an example of a json message that is submitting a text legacy governance: (In this example, the `proposer` is the address of the interchain account on the host chain)

```json
{
  "messages": [
    {
      "@type": "/cosmos.gov.v1beta1.MsgSubmitProposal",
      "content": {
        "@type": "/cosmos.gov.v1beta1.TextProposal",
        "title": "IBC Gov Proposal",
        "description": "tokens for all!"
      },
      "initial_deposit": [{ "denom": "stake", "amount": "5000" }],
      "proposer": "cosmos1k4epd6js8aa7fk4e5l7u6dwttxfarwu6yald9hlyckngv59syuyqnlqvk8"
    }
  ]
}
```

If the channel is using `proto3` (protobuf) encoding, then the format that protobuf messages have to take are defined by the cosmos-sdk's protobuf codec. Protobuf messages do not have nice human readable formats like json messages do. In the rust the `cosmos-sdk-proto` library is used to generate the protobuf messages as follows:

```rust
use cosmos_sdk_proto::{
    cosmos::{bank::v1beta1::MsgSend, base::v1beta1::Coin},
    traits::MessageExt,
};

// ...

let predefined_proto_message = MsgSend {
    from_address: ica_info.ica_address,
    to_address,
    amount: vec![Coin {
        denom: "stake".to_string(),
        amount: "100".to_string(),
    }],
};

IcaPacketData::from_proto_anys(predefined_proto_message.to_any()?);
```

where `from_proto_anys` is defined as:

```rust
pub use cosmos_sdk_proto::ibc::applications::interchain_accounts::v1::CosmosTx;

// ...

/// Creates a new IcaPacketData from a list of [`cosmos_sdk_proto::Any`] messages
pub fn from_proto_anys(messages: Vec<cosmos_sdk_proto::Any>, memo: Option<String>) -> Self {
    let cosmos_tx = CosmosTx { messages };
    let data = cosmos_tx.encode_to_vec();
    Self::new(data, memo)
}
```

And in golang [(see e2e tests)](./e2e/interchaintest/types/contract_msg.go) we use the `SerializeCosmosTxWithEncoding` from ibc-go to encode the protobuf messages where encoding is either `proto3` or `proto3json`:

```go
// NewSendCustomIcaMessagesMsg creates a new SendCustomIcaMessagesMsg.
func NewSendCustomIcaMessagesMsg(cdc codec.BinaryCodec, msgs []proto.Message, encoding string, memo *string, timeout *uint64) string {
	type SendCustomIcaMessagesMsg struct {
		Messages       string  `json:"messages"`
		PacketMemo     *string `json:"packet_memo,omitempty"`
		TimeoutSeconds *uint64 `json:"timeout_seconds,omitempty"`
	}

	type SendCustomIcaMessagesMsgWrapper struct {
		SendCustomIcaMessagesMsg SendCustomIcaMessagesMsg `json:"send_custom_ica_messages"`
	}

	bz, err := icatypes.SerializeCosmosTxWithEncoding(cdc, msgs, encoding)
	if err != nil {
		panic(err)
	}

	messages := base64.StdEncoding.EncodeToString(bz)

	msg := SendCustomIcaMessagesMsgWrapper{
		SendCustomIcaMessagesMsg: SendCustomIcaMessagesMsg{
			Messages:       messages,
			PacketMemo:     memo,
			TimeoutSeconds: timeout,
		},
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
```

2. `SendPredefinedAction`: This message sends a 100 stake from the ica account to a user defined address on the host chain. This action is used to demonstrate how you can have a contract that executes a predefined action on the host chain. This is more useful for DAOs or other contracts that need to execute specific actions on the host chain. This message type checks which encoding the channel is using, and sends the appropriate message to the host chain. The host chain then executes the message, and sends the result back to this contract. A callback is then executed based on the result.

### Execute a callback

This contract supports callbacks. See [`src/ibc/relay.rs`](./src/ibc/relay.rs) to learn how to decode whether a transaction was successful or not. Currently, a counter is incremented to record how many transactions were successful and how many failed. This is just a placeholder for more complex logic that can be executed in the callback.

### Channel Closing and Reopening

In this contract, the only way to close the ica channel is to timeout the channel (you can see this in the end to end tests). This is because the semantics of ordered channels in IBC is that any timeout will cause the channel to be closed. The contract is then able to create a new channel with the same interchain account address, and continue to use the same interchain account. This can also be seen in the end to end tests.

## Testing

There are two kinds of tests in this repository: unit tests and end to end tests. The unit tests are located inside the rust files in the `src` directory. The end to end tests are located in the `e2e` directory.

### Unit tests

In general, the unit tests are for testing the verification functions for the handshake, and for testing that the serializers and deserializers are working correctly. To run the unit tests, run `cargo test`.

### End to end tests

The end to end tests are for testing the contract's functionality in an environment mimicking production. To see whether or not it can perform the channel handshake, send packets, and execute callbacks. We achieve this by running two local chains, one for the contract, and one for the host chain. The relayer is then used to perform the channel handshake, and send packets. The contract then executes callbacks based on the result of the packet. To learn more about how to run the end to end tests, see the [Readme](./e2e/Readme.md) in the `e2e` directory.

## Limitations

This contract is not meant to be used in production. It is meant to be used as a reference implementation for how to build a CosmWasm contract that can communicate with the golang ica/host module. The following are some of the limitations of this contract:

- The contract cannot create multiple interchain accounts. It can only create one.
- ICA channels must be ordered (enforced by golang ica/host module). Due to the semantics of ordered channels in IBC, any timeout will cause the channel to be closed.
- The relayer must start the channel handshake on the contract's chain. This is not possible to do in the contract itself. See e2e tests for an example of how to do this.
- The contract cannot initialize with an empty string as the version. This is due to a limitation of the IBCModule interface provided by ibc-go, see issue [#3942](https://github.com/cosmos/ibc-go/issues/3942). (The contract can be initialized with an empty version string if the chain supports stargate queries, but this is not the case for the end to end tests.)

## Acknowledgements

Much thanks to [Art3mix](https://github.com/Art3miX) for all the helpful discussions and nailing down of the encoding/decoding issues. Also thanks to [0xekez](https://github.com/0xekez) for their work on [cw-ibc-example](https://github.com/0xekez/cw-ibc-example) which was a great reference for CosmWasm IBC endpoints and interchaintest.
