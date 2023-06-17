use cosmwasm_schema::{cw_serde, QueryResponses};

use crate::types::state::{ContractChannelState, ContractState};

#[cw_serde]
pub struct InstantiateMsg {
    pub admin: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(ContractChannelState)]
    GetChannel {},
    #[returns(ContractState)]
    GetContractState {},
}
