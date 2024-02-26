//! # Messages
//!
//! This module defines the messages that this contract receives.

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Binary, CosmosMsg};

/// The message to instantiate the ICA controller contract.
#[cw_serde]
pub struct InstantiateMsg {
    /// The address of the owner of the ICA application.
    /// If not specified, the sender is the owner.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub owner: Option<String>,
    /// The options to initialize the IBC channel upon contract instantiation.
    pub channel_open_init_options: options::ChannelOpenInitOptions,
    /// The contract address that the channel and packet lifecycle callbacks are sent to.
    /// If not specified, then no callbacks are sent.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub send_callbacks_to: Option<CallbackInfo>,
}

/// The info needed to send callbacks
#[cw_serde]
pub struct CallbackInfo {
    /// The address of the callback contract.
    pub address: String,
    /// The code hash of the callback contract.
    pub code_hash: String,
}

/// The messages to execute the ICA controller contract.
#[cw_serde]
pub enum ExecuteMsg {
    /// `CreateChannel` makes the contract submit a stargate MsgChannelOpenInit to the chain.
    /// This is a wrapper around [`options::ChannelOpenInitOptions`] and thus requires the
    /// same fields. If not specified, then the options specified in the contract instantiation
    /// are used.
    CreateChannel {
        /// The options to initialize the IBC channel.
        /// If not specified, the options specified in the last channel creation are used.
        /// Must be `None` if the sender is not the owner.
        #[serde(skip_serializing_if = "Option::is_none")]
        channel_open_init_options: Option<options::ChannelOpenInitOptions>,
    },
    /// `CloseChannel` closes the IBC channel.
    CloseChannel {},
    /// `SendCosmosMsgs` converts the provided array of [`CosmosMsg`] to an ICA tx and sends them to the ICA host.
    /// [`CosmosMsg::Stargate`] and [`CosmosMsg::Wasm`] are only supported if the [`TxEncoding`](crate::ibc::types::metadata::TxEncoding) is [`TxEncoding::Protobuf`](crate::ibc::types::metadata::TxEncoding).
    ///
    /// **This is the recommended way to send messages to the ICA host.**
    SendCosmosMsgs {
        /// The stargate messages to convert and send to the ICA host.
        messages: Vec<CosmosMsg>,
        /// Optional memo to include in the ibc packet.
        #[serde(skip_serializing_if = "Option::is_none")]
        packet_memo: Option<String>,
        /// Optional timeout in seconds to include with the ibc packet.
        /// If not specified, the [default timeout](crate::ibc::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
        #[serde(skip_serializing_if = "Option::is_none")]
        timeout_seconds: Option<u64>,
    },
    /// `SendCustomIcaMessages` sends custom messages from the ICA controller to the ICA host.
    ///
    /// **Use this only if you know what you are doing.**
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
        /// If not specified, the [default timeout](crate::ibc::types::packet::DEFAULT_TIMEOUT_SECONDS) is used.
        #[serde(skip_serializing_if = "Option::is_none")]
        timeout_seconds: Option<u64>,
    },
    /// `UpdateCallbackAddress` updates the contract callback address.
    UpdateCallbackAddress {
        /// The new callback address.
        /// If not specified, then no callbacks are sent.
        callback_contract: Option<CallbackInfo>,
    },
    /// `UpdateOwnership` updates the contract owner.
    UpdateOwnership {
        /// The new owner of the contract.
        owner: String,
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
    /// Ownership returns the owner of the contract.
    #[returns(String)]
    Ownership {},
}

/// Option types for other messages.
pub mod options {
    use cosmwasm_std::IbcOrder;

    use super::cw_serde;
    use crate::ibc::types::{keys::HOST_PORT_ID, metadata::TxEncoding};

    /// The message used to provide the MsgChannelOpenInit with the required data.
    #[cw_serde]
    pub struct ChannelOpenInitOptions {
        /// The connection id on this chain.
        pub connection_id: String,
        /// The counterparty connection id on the counterparty chain.
        pub counterparty_connection_id: String,
        /// The counterparty port id. If not specified, [`crate::ibc::types::keys::HOST_PORT_ID`] is used.
        /// Currently, this contract only supports the host port.
        pub counterparty_port_id: Option<String>,
        /// TxEncoding is the encoding used for the ICA txs. If not specified, [`TxEncoding::Protobuf`] is used.
        pub tx_encoding: Option<TxEncoding>,
        /// The order of the channel. If not specified, [`IbcOrder::Ordered`] is used.
        /// [`IbcOrder::Unordered`] is only supported if the counterparty chain is using `ibc-go`
        /// v8.1.0 or later.
        pub channel_ordering: Option<IbcOrder>,
    }

    impl ChannelOpenInitOptions {
        /// Returns the counterparty port id.
        #[must_use]
        pub fn counterparty_port_id(&self) -> String {
            self.counterparty_port_id
                .clone()
                .unwrap_or_else(|| HOST_PORT_ID.to_string())
        }

        /// Returns the tx encoding.
        #[must_use]
        pub fn tx_encoding(&self) -> TxEncoding {
            self.tx_encoding.clone().unwrap_or(TxEncoding::Protobuf)
        }
    }
}
