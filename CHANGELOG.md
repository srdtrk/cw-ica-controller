# Changelog

## v0.1.1 (2023-10-21)

Initial release.

### Features

- Relayer initiated channel opening.
- Added `ExecuteMsg::SendCustomIcaMessages` to send custom ICA messages.
- Added `ExecuteMsg::SendPredefinedAction` to send predefined ICA messages for testing.
- Added a `CallbackCounter` to count the number of ICA callbacks.

## v0.1.2 (2023-10-25)

### Features

- Added `helpers.rs` for external contracts. (https://github.com/srdtrk/cw-ica-controller/commit/bef4d34cc674892725c36c5fcbb467aaa38c38c8)

## v0.1.3 (2023-10-28)

### Features

- Added contract instantiated channel opening. (https://github.com/srdtrk/cw-ica-controller/pull/13)

## v0.2.0 (2023-11-09)

### Features

- Added callbacks to external contracts. (https://github.com/srdtrk/cw-ica-controller/pull/16)

### API Breaking Changes

- Removed `ExecuteMsg::SendPredefinedAction` (https://github.com/srdtrk/cw-ica-controller/pull/16)
- Removed `library` feature. (https://github.com/srdtrk/cw-ica-controller/pull/20)
