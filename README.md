# CosmWasm ICA Controller Contract

This is a CosmWasm smart contract that communicates with the golang ica/host module on the host chain to create and manage one interchain account. This contract can also execute callbacks based on the result of the interchain account transaction. Because this is a CosmWasm implementation of the entire ICA controller, the chain that this contract is deployed on need **not** have the ICA module enabled. This contract can be deployed on any chain that supports IBC CosmWasm smart contracts.

This contract was originally written to test the json encoding and decoding [feature being added to interchain accounts](https://github.com/cosmos/ibc-go/pull/3796). Therefore, this contract cannot function in mainnet until this PR is merged, and backported to the version of ibc-go used by the host chain.

Note that the same approach used to build this contract can be used to build a contract that works in mainnet today, as long as the correct protobuf messages are used.

## Usage

You can see the various ways this contract can be used in the end to end tests in the `e2e` directory. The following is a brief overview of the contract's functionality.

### Create an interchain account

To create an interchain account, the relayer must start the channel handshake on the contract's chain. See end to end tests for an example of how to do this. Unfortunately, this is not possible to do in the contract itself. Also, you cannot initialize with an empty string as the version, this is due to a limitation of the IBCModule interface provided by ibc-go, see issue [#3942](https://github.com/cosmos/ibc-go/issues/3942). The version string we are using for tests is: `{"version":"ics27-1","controller_connection_id":"connection-0","host_connection_id":"connection-0","address":"","encoding":"proto3json","tx_type":"sdk_multi_msg"}`. You can see this in the end to end tests.

### Execute an interchain account transaction

In this contract, the `execute` message is used to commit a packet to be sent to the host chain. This contract has two ways of executing an interchain transaction:

1. `SendCustomIcaMessages`: This message requires the sender to give json/base64 encoded messages that will be sent to the host chain. The host chain will decode the messages and execute them. The result of the execution will be sent back to this contract, and the callback will be executed.

The format that json messages have to take are defined by the cosmos-sdk's json codec. The following is an example of a json message that is submitting a text legacy governance: (In this example, the `proposer` is the address of the interchain account on the host chain)

```json
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
```

2. `SendPredefinedAction`: This message sends a 100 stake from the ica account to a user defined address on the host chain. This action is used to demonstrate how you can have a contract that executes a predefined action on the host chain. This is more useful for DAOs or other contracts that need to execute specific actions on the host chain.

### Execute a callback

This contract supports callbacks. See `src/ibc/relay.rs` for how to decode whether a transaction was successful or not. Currently, a counter is incremented to record how many transactions were successful and how many failed. This is just a placeholder for more complex logic that can be executed in the callback.

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
- Channel re-opening is not supported (yet).
- The relayer must start the channel handshake on the contract's chain. This is not possible to do in the contract itself. See e2e tests for an example of how to do this.
- The contract cannot initialize with an empty string as the version. This is due to a limitation of the IBCModule interface provided by ibc-go, see issue [#3942](https://github.com/cosmos/ibc-go/issues/3942).
