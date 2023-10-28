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
    /// The options to initialize the IBC channel upon contract instantiation.
    /// If not specified, the IBC channel is not initialized, and the relayer must.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub channel_open_init_options: Option<options::ChannelOpenInitOptions>,
}

/// The messages to execute the ICA controller contract.
#[cw_serde]
pub enum ExecuteMsg {
    /// CreateChannel makes the contract submit a stargate MsgChannelOpenInit to the chain.
    /// This is a wrapper around [`options::ChannelOpenInitOptions`] and thus requires the
    /// same fields.
    CreateChannel(options::ChannelOpenInitOptions),
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

/// Option types for other messages.
pub mod options {
    use super::*;
    use crate::ibc::types::{keys::HOST_PORT_ID, metadata::TxEncoding};

    /// The message used to provide the MsgChannelOpenInit with the required data.
    #[cw_serde]
    pub struct ChannelOpenInitOptions {
        /// The connection id on this chain.
        pub connection_id: String,
        /// The counterparty connection id on the counterparty chain.
        pub counterparty_connection_id: String,
        /// The counterparty port id. If not specified, [crate::ibc::types::keys::HOST_PORT_ID] is used.
        /// Currently, this contract only supports the host port.
        pub counterparty_port_id: Option<String>,
        /// TxEncoding is the encoding used for the ICA txs. If not specified, [TxEncoding::Protobuf] is used.
        pub tx_encoding: Option<TxEncoding>,
    }

    impl ChannelOpenInitOptions {
        /// Returns the counterparty port id.
        pub fn counterparty_port_id(&self) -> String {
            self.counterparty_port_id
                .clone()
                .unwrap_or(HOST_PORT_ID.to_string())
        }

        /// Returns the tx encoding.
        pub fn tx_encoding(&self) -> TxEncoding {
            self.tx_encoding.clone().unwrap_or(TxEncoding::Protobuf)
        }
    }
}
