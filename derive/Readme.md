# `cw-ica-controller-derive` - ICA Controller Derive

This crate provides a derive macro for contracts receiving ICA controller callback messages.
This crate's macros are not intended to be used directly, but rather as a dependency of the
`cw-ica-controller` crate where it is re-exported under the `cw_ica_controller::helpers`.

This allows the users of the `cw-ica-controller` crate to easily merge the required callback
message enum variant into their `ExecuteMsg` enum.

## Usage

I will show the usage of this crate (from the `cw-ica-controller` crate) in
[`testing/contracts/callback-counter/src/msg.rs`](../testing/contracts/callback-counter/src/msg.rs).

```rust
use cosmwasm_schema::{cw_serde, QueryResponses};
use cw_ica_controller::helpers::ica_callback_execute;

#[cw_serde]
pub struct InstantiateMsg {}

#[ica_callback_execute]
#[cw_serde]
pub enum ExecuteMsg {}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetCallbackCounter returns the callback counter.
    #[returns(crate::state::CallbackCounter)]
    GetCallbackCounter {},
}
```
