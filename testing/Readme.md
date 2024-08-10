# Testing Contracts

This directory contains the smart contracts used in the [end-to-end testing](../e2e/Readme.md) of the `cw-ica-controller`. This directory is not intended to be used in production.

## Contracts

### `callback-counter`

This contract is solely used to receive and count callbacks from the `cw-ica-controller`. It is used to test the `cw-ica-controller`'s ability to callback to a contract.
It is also used to test whether or not each ICA tx returns the expected type of callback. (`success`, `failure`, or `timeout`)

### `cw-ica-owner`

This contract is used to test how the `cw-ica-controller` could be controlled by another smart contract. It is used to test the `cw-ica-controller`'s API.

## Building the Contracts

The test contracts are built in the github CI. To build the contracts manually, run the following command:

```sh
just build-test-contracts
```
