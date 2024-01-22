use cosmwasm_schema::{cw_serde, QueryResponses};
use cw_ica_controller::types::callbacks::IcaControllerCallbackMsg;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    ReceiveIcaCallback(IcaControllerCallbackMsg),
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetCallbackCounter returns the callback counter.
    #[returns(crate::state::CallbackCounter)]
    GetCallbackCounter {},
}
