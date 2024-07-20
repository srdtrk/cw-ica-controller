# End to End Tests

The e2e tests are built using the [interchaintest](https://github.com/strangelove-ventures/interchaintest) library by Strangelove. It runs multiple docker container validators, and lets you test IBC enabled smart contracts.

These end to end tests are designed to run in the ci, but you can also run them locally.

## Running the tests locally

The end to end tests are currently split into two parts:

### ICA Contract Tests

These tests are designed to test the ICA contract itself and its interaction with the relayer.

All contract tests are located in `interchaintest/contract_test.go` file. Currently, there are four tests in this file:

- TestWithContractTestSuite/TestIcaContractChannelHandshake_Ordered_Protobuf
- TestWithContractTestSuite/TestIcaContractChannelHandshake_Unordered_Protobuf
- TestWithContractTestSuite/TestIcaRelayerInstantiatedChannelHandshake
- TestWithContractTestSuite/TestRecoveredIcaContractInstantiatedChannelHandshake
- TestWithContractTestSuite/TestIcaContractExecution_Ordered_Protobuf
- TestWithContractTestSuite/TestIcaContractExecution_Unordered_Protobuf
- TestWithContractTestSuite/TestIcaContractTimeoutPacket_Ordered_Protobuf
- TestWithContractTestSuite/TestIcaContractTimeoutPacket_Unordered_Protobuf
- TestWithOwnerTestSuite/TestOwnerCreateIcaContract
- TestWithOwnerTestSuite/TestOwnerPredefinedAction
- TestWithContractTestSuite/TestSendCosmosMsgs_Ordered_Protobuf
- TestWithContractTestSuite/TestSendCosmosMsgs_Unordered_Protobuf
- TestWithWasmTestSuite/TestSendWasmMsgs
- TestWithContractTestSuite/TestMigrateOrderedToUnordered
- TestWithContractTestSuite/TestCloseChannel_Protobuf_Unordered
- TestWithContractTestSuite/TestBankAndStargateQueries
- TestWithContractTestSuite/TestStakingQueries

To run the tests locally, run the following command in the root of the repository:

```sh
just e2e-test $TEST_NAME
```

where `$TEST_NAME` is one of the tests listed above.

Before running the tests, you must have built the optimized contract in the `/artifacts` directory. To do this, run the following commands from the root of the repository:

```sh
just build-optimized
just build-test-contracts
```

## In the CI

The tests are run in the github CI after every push to the `main` branch and in every PR. See the [github actions workflow](https://github.com/srdtrk/cw-ica-controller/blob/main/.github/workflows/e2e.yml) for more details.

## About the tests

The tests are currently run on wasmd `v0.50.0` and ibc-go `v8.1.0`.
