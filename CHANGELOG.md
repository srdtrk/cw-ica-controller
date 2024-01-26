# Changelog

## v0.4.0 (2024-01-27)

### Features

- `DistributionMsg::FundCommunityPool` is now supported in `ExecuteMsg::SendCosmosMsgs`. (https://github.com/srdtrk/cw-ica-controller/pull/46)

### Breaking Changes

- `InstantiateMsg`'s `channel_open_init_options` field is now required. (https://github.com/srdtrk/cw-ica-controller/pull/53)
- `CallbackCounter` is removed. (https://github.com/srdtrk/cw-ica-controller/pull/44)
- Relayers cannot open channels anymore. (https://github.com/srdtrk/cw-ica-controller/pull/53)
- Minimum compatible CosmWasm version is now `v1.3`. (https://github.com/srdtrk/cw-ica-controller/pull/46)
- Improved the `instantiate2` helper function. (https://github.com/srdtrk/cw-ica-controller/pull/57)

## v0.3.0 (2023-12-30)

### Features

- Added `CosmosMsg` support. (https://github.com/srdtrk/cw-ica-controller/pull/28)

### API Breaking Changes

- Removed stargate query fallback in the contract. (https://github.com/srdtrk/cw-ica-controller/pull/31)
- Switched to `cw-ownable` for contract's admin management. (https://github.com/srdtrk/cw-ica-controller/pull/25)

## v0.2.0 (2023-11-09)

### Features

- Added callbacks to external contracts. (https://github.com/srdtrk/cw-ica-controller/pull/16)

### API Breaking Changes

- Removed `ExecuteMsg::SendPredefinedAction` (https://github.com/srdtrk/cw-ica-controller/pull/16)
- Removed `library` feature. (https://github.com/srdtrk/cw-ica-controller/pull/20)

## v0.1.3 (2023-10-28)

### Features

- Added contract instantiated channel opening. (https://github.com/srdtrk/cw-ica-controller/pull/13)

## v0.1.2 (2023-10-25)

### Features

- Added `helpers.rs` for external contracts. (https://github.com/srdtrk/cw-ica-controller/commit/bef4d34cc674892725c36c5fcbb467aaa38c38c8)

## v0.1.1 (2023-10-21)

Initial release.

### Features

- Relayer initiated channel opening.
- Added `ExecuteMsg::SendCustomIcaMessages` to send custom ICA messages.
- Added `ExecuteMsg::SendPredefinedAction` to send predefined ICA messages for testing.
- Added a `CallbackCounter` to count the number of ICA callbacks.
