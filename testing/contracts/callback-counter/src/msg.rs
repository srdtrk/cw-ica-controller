use cosmwasm_schema::{cw_serde, QueryResponses};
use cw_ica_controller::helpers::ica_callback_execute;

#[cw_serde]
pub struct InstantiateMsg {}

#[ica_callback_execute]
#[deny(missing_docs)]
#[cw_serde]
/// This is the execute message of the contract.
pub enum ExecuteMsg {}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetCallbackCounter returns the callback counter.
    #[returns(crate::state::CallbackCounter)]
    GetCallbackCounter {},
}
