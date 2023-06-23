# End to End Tests

The e2e tests are built using the [interchaintest](https://github.com/strangelove-ventures/interchaintest) library by Strangelove. It runs multiple docker container validators, and lets you test IBC enabled smart contracts.

## Running the tests

```text
cd interchaintest/
go test -v contract_test.go
```

## In the CI

The tests are run in the github CI after every push to the `main` branch. See the [github actions workflow](https://github.com/srdtrk/cw-ica-controller/blob/main/.github/workflows/e2e.yml) for more details.

## About the tests

The tests are currently run on wasmd `v0.40.2` and ibc-go's simd `pr-3796` which implements json encoding for interchain accounts module. Once this pr is merged, we will update the tests to use the latest version of ibc-go.
