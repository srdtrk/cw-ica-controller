# Changelog

## [Unreleased]

### API Breaking Changes

- Removed support for `proto3json` encoding. (https://github.com/srdtrk/cw-ica-controller/pull/92)
- Removed `tx_encoding` field from `ChannelOpenInitOptions`. (https://github.com/srdtrk/cw-ica-controller/pull/92)
- Removed `ExecuteMsg::SendCustomIcaMessages`. (https://github.com/srdtrk/cw-ica-controller/pull/92)

## v0.5.0 (2024-02-05)

### Features

- Added support for UNORDERED channels introduced to `icahost` in `ibc-go` v8.1.0. (https://github.com/srdtrk/cw-ica-controller/pull/74)
- Added `ExecuteMsg::CloseChannel` to close a channel so that it may be reopened with different options. (https://github.com/srdtrk/cw-ica-controller/pull/78)

### API Breaking Changes

- Removed `allow_channel_open_init` from `ContractState`. (https://github.com/srdtrk/cw-ica-controller/pull/76)
- Removed needless pass by value in `helpers.rs`. (https://github.com/srdtrk/cw-ica-controller/pull/76)
- Added `channel_ordering` field to `ChannelOpenInitOptions`. (https://github.com/srdtrk/cw-ica-controller/pull/74)

## v0.4.2 (2024-01-28)

### Changes

- Added inline documentation to the enum variant inserted by the proc macro introduced in `v0.4.1`. (https://github.com/srdtrk/cw-ica-controller/pull/66)

## v0.4.1 (2024-01-27)

### Features

- Introduced a proc macro to insert the callback msg enum variant to external contracts' `ExecuteMsg`. (https://github.com/srdtrk/cw-ica-controller/pull/61)

## v0.4.0 (2024-01-27)

### Features

- `DistributionMsg::FundCommunityPool` is now supported in `ExecuteMsg::SendCosmosMsgs`. (https://github.com/srdtrk/cw-ica-controller/pull/46)

### Breaking Changes

- `InstantiateMsg`'s `channel_open_init_options` field is now required. (https://github.com/srdtrk/cw-ica-controller/pull/53)
- `CallbackCounter` is removed. (https://github.com/srdtrk/cw-ica-controller/pull/44)
- Relayers cannot initiate opening channels anymore. (https://github.com/srdtrk/cw-ica-controller/pull/53)
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
