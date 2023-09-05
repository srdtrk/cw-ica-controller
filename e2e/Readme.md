# End to End Tests

The e2e tests are built using the [interchaintest](https://github.com/strangelove-ventures/interchaintest) library by Strangelove. It runs multiple docker container validators, and lets you test IBC enabled smart contracts.

## Running the tests locally

All contract tests are located in `interchaintest/contract_test.go` file. Currently, there are four tests in this file:

- `TestIcaContractChannelHandshake`
- `TestIcaContractExecutionProto3JsonEncoding`
- `TestIcaContractExecutionProtobufEncoding`
- `TestIcaContractTimeoutPacket`

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

## In the CI

The tests are run in the github CI after every push to the `main` branch. See the [github actions workflow](https://github.com/srdtrk/cw-ica-controller/blob/main/.github/workflows/e2e.yml) for more details.

For some unknown reason, the timeout test sometimes fails in the CI (I'd say about 20-25% of the time). In this case, feel free to rerun the CI job.

## About the tests

The tests are currently run on wasmd `v0.40.2` and ibc-go `v7.3.0`'s simd which implements json encoding feature for the interchain accounts module.
