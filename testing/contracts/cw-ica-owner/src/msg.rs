use cosmwasm_schema::{cw_serde, QueryResponses};
use cw_ica_controller::types::msg::options::ChannelOpenInitOptions;

#[cw_serde]
pub struct InstantiateMsg {
    pub admin: Option<String>,
    pub ica_controller_code_id: u64,
}

#[cw_serde]
pub enum ExecuteMsg {
    CreateIcaContract {
        salt: Option<String>,
        channel_open_init_options: Option<ChannelOpenInitOptions>,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetContractState returns the contact's state.
    #[returns(crate::state::ContractState)]
    GetContractState {},
    /// GetIcaState returns the ICA state for the given ICA ID.
    #[returns(crate::state::IcaContractState)]
    GetIcaContractState { ica_id: u64 },
    /// GetIcaCount returns the number of ICAs.
    #[returns(u64)]
    GetIcaCount {},
}
