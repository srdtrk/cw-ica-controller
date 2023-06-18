use cosmwasm_schema::{cw_serde, QueryResponses};

use crate::types::state::{ChannelState, ContractState};

#[cw_serde]
pub struct InstantiateMsg {
    pub admin: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    /// SendCustomIcaMessages sends custom messages from the ICA controller to the ICA host.
    /// It works by replacing `"$ICA_ADDRESS"` in each message with the ICA address.
    SendCustomIcaMessages {
        messages: Vec<String>,
        packet_memo: Option<String>,
        timeout_seconds: Option<u64>,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(ChannelState)]
    GetChannel {},
    #[returns(ContractState)]
    GetContractState {},
}
