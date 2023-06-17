use cosmwasm_schema::{cw_serde, QueryResponses};

use crate::types::state::{ChannelState, ContractState};

#[cw_serde]
pub struct InstantiateMsg {
    pub admin: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(ChannelState)]
    GetChannel {},
    #[returns(ContractState)]
    GetContractState {},
}
