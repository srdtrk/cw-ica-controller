# End to End Tests

The e2e tests are built using the [interchaintest](https://github.com/strangelove-ventures/interchaintest) library by Strangelove. It runs multiple docker container validators, and lets you test IBC enabled smart contracts.

These end to end tests are designed to run in the ci, but you can also run them locally.

## Running the tests locally

The end to end tests are currently split into two parts:

### ICA Contract Tests

These tests are designed to test the ICA contract itself and its interaction with the relayer.

All contract tests are located in `interchaintest/contract_test.go` file. Currently, there are four tests in this file:

- `TestIcaContractChannelHandshake_Ordered_Protobuf`
- `TestIcaContractChannelHandshake_Unordered_Protobuf`
- `TestIcaRelayerInstantiatedChannelHandshake`
- `TestRecoveredIcaContractInstantiatedChannelHandshake`
- `TestIcaContractExecution_Ordered_Protobuf`
- `TestIcaContractExecution_Unordered_Protobuf`
- `TestIcaContractTimeoutPacket_Ordered_Protobuf`
- `TestIcaContractTimeoutPacket_Unordered_Protobuf`
- `TestSendCosmosMsgs_Ordered_Protobuf`
- `TestSendCosmosMsgs_Unordered_Protobuf`
- `TestSendWasmMsgsProtobufEncoding`
- `TestMigrateOrderedToUnordered`
- `TestCloseChannel_Protobuf_Unordered`

(These three tests used to be one monolithic test, but they were split into three in order to run them in parallel in the CI.)

To run the tests locally, run the following commands from this directory:

```text
cd interchaintest/
go test -v . -run TestWithContractTestSuite -testify.m $TEST_NAME
```

where `$TEST_NAME` is one of the four tests listed above.

Before running the tests, you must have built the optimized contract in the `/artifacts` directory. To do this, run the following command from the root of the repository:

```text
cargo run-script optimize
```

### Owner Contract Tests

These tests are designed to test the ICA contract's interaction with external contracts such as callbacks. For this, a mock owner contract is used.

All owner contract tests are located in `interchaintest/owner_test.go` file. Currently, there are two tests in this file:

- `TestOwnerCreateIcaContract`
- `TestOwnerPredefinedAction`

```text
cd interchaintest/
go test -v . -run TestWithOwnerTestSuite -testify.m $TEST_NAME
```

where `$TEST_NAME` is one of the two tests listed above.

## In the CI

The tests are run in the github CI after every push to the `main` branch. See the [github actions workflow](https://github.com/srdtrk/cw-ica-controller/blob/main/.github/workflows/e2e.yml) for more details.

For some unknown reason, the timeout test sometimes fails in the CI (I'd say about 20-25% of the time). In this case, feel free to rerun the CI job.

## About the tests

The tests are currently run on wasmd `v0.41.0` and ibc-go `v7.3.0`'s simd which implements json encoding feature for the interchain accounts module.
