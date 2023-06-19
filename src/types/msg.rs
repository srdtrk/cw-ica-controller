use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;

use crate::types::state::{CallbackCounter, ChannelState, ContractState};

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
        ///
        /// # Example Message:
        ///
        /// This is a legacy text governance proposal message.
        ///
        /// ```json
        /// {
        ///   "@type": "/cosmos.gov.v1beta1.MsgSubmitProposal",
        ///   "content": {
        ///     "@type": "/cosmos.gov.v1beta1.TextProposal",
        ///     "title": "IBC Gov Proposal",
        ///     "description": "tokens for all!"
        ///   },
        ///   "initial_deposit": [{ "denom": "stake", "amount": "5000" }],
        ///   "proposer": "$ica_address"
        /// }
        /// ```
        ///
        /// `$ica_address` will be replaced with the ICA address before the message is sent to the ICA host.
        messages: Vec<Binary>,
        /// Optional memo to include in the ibc packet.
        packet_memo: Option<String>,
        /// Optional timeout in seconds to include with the ibc packet.
        /// If not specified, the [default timeout](crate::ibc_module::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
        timeout_seconds: Option<u64>,
    },
    /// SendPredefinedAction sends a predefined action from the ICA controller to the ICA host.
    /// This demonstration is useful for contracts that have predefined actions such as DAOs.
    ///
    /// In this example, the predefined action is a `MsgSend` message which sends 100 "stake" tokens.
    SendPredefinedAction { to_address: String },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(ChannelState)]
    GetChannel {},
    #[returns(ContractState)]
    GetContractState {},
    #[returns(CallbackCounter)]
    GetCallbackCounter {},
}
