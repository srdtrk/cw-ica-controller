//! # Messages
//!
//! This module defines the messages the ICA controller contract receives.

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;

/// The message to instantiate the ICA controller contract.
#[cw_serde]
pub struct InstantiateMsg {
    /// The address of the admin of the ICA application.
    /// If not specified, the sender is the admin.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub admin: Option<String>,
}

/// The messages to execute the ICA controller contract.
#[cw_serde]
pub enum ExecuteMsg {
    /// SendCustomIcaMessages sends custom messages from the ICA controller to the ICA host.
    SendCustomIcaMessages {
        /// Base64-encoded json or proto messages to send to the ICA host.
        ///
        /// # Example JSON Message:
        ///
        /// This is a legacy text governance proposal message serialized using proto3json.
        ///
        /// ```json
        ///  {
        ///    "messages": [
        ///      {
        ///        "@type": "/cosmos.gov.v1beta1.MsgSubmitProposal",
        ///        "content": {
        ///          "@type": "/cosmos.gov.v1beta1.TextProposal",
        ///          "title": "IBC Gov Proposal",
        ///          "description": "tokens for all!"
        ///        },
        ///        "initial_deposit": [{ "denom": "stake", "amount": "5000" }],
        ///        "proposer": "cosmos1k4epd6js8aa7fk4e5l7u6dwttxfarwu6yald9hlyckngv59syuyqnlqvk8"
        ///      }
        ///    ]
        ///  }
        /// ```
        ///
        /// where proposer is the ICA controller's address.
        messages: Binary,
        /// Optional memo to include in the ibc packet.
        #[serde(skip_serializing_if = "Option::is_none")]
        packet_memo: Option<String>,
        /// Optional timeout in seconds to include with the ibc packet.
        /// If not specified, the [default timeout](crate::ibc_module::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
        #[serde(skip_serializing_if = "Option::is_none")]
        timeout_seconds: Option<u64>,
    },
    /// SendPredefinedAction sends a predefined action from the ICA controller to the ICA host.
    /// This demonstration is useful for contracts that have predefined actions such as DAOs.
    ///
    /// In this example, the predefined action is a `MsgSend` message which sends 100 "stake" tokens.
    SendPredefinedAction {
        /// The recipient's address, on the counterparty chain, to send the tokens to from ICA host.
        to_address: String,
    },
}

/// The messages to query the ICA controller contract.
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetChannel returns the IBC channel info.
    #[returns(crate::types::state::ChannelState)]
    GetChannel {},
    /// GetContractState returns the contact's state.
    #[returns(crate::types::state::ContractState)]
    GetContractState {},
    /// GetCallbackCounter returns the callback counter.
    #[returns(crate::types::state::CallbackCounter)]
    GetCallbackCounter {},
}

/// The message to migrate this contract.
#[cw_serde]
pub struct MigrateMsg {}
