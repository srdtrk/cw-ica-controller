use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;

use crate::types::state::{ChannelState, ContractState};

#[cw_serde]
pub struct InstantiateMsg {
    pub admin: Option<String>,
}

#[cw_serde]
pub enum ExecuteMsg {
    /// SendCustomIcaMessages sends custom messages from the ICA controller to the ICA host.
    /// It works by replacing [`ICA_PLACEHOLDER`](crate::types::keys::ICA_PLACEHOLDER) in each message with the ICA address.
    SendCustomIcaMessages {
        /// Base64-encoded json messages to send to the ICA host.
        messages: Vec<Binary>,
        /// Optional memo to include with the ibc packet.
        packet_memo: Option<String>,
        /// Optional timeout in seconds to include with the ibc packet.
        /// If not specified, the default timeout is used.
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
