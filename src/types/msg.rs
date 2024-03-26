//! # Messages
//!
//! This module defines the messages that this contract receives.

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::CosmosMsg;

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
    pub send_callbacks_to: Option<String>,
}

/// The messages to execute the ICA controller contract.
#[cw_ownable::cw_ownable_execute]
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
    /// `UpdateCallbackAddress` updates the contract callback address.
    UpdateCallbackAddress {
        /// The new callback address.
        /// If not specified, then no callbacks are sent.
        callback_address: Option<String>,
    },
}

/// The messages to query the ICA controller contract.
#[cw_ownable::cw_ownable_query]
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// GetChannel returns the IBC channel info.
    #[returns(crate::types::state::ChannelState)]
    GetChannel {},
    /// GetContractState returns the contact's state.
    #[returns(crate::types::state::ContractState)]
    GetContractState {},
}

/// The message to migrate this contract.
#[cw_serde]
pub struct MigrateMsg {}

/// Option types for other messages.
pub mod options {
    use cosmwasm_std::IbcOrder;

    /// The options needed to initialize the IBC channel.
    #[derive(
        serde::Serialize,
        serde::Deserialize,
        Clone,
        Debug,
        PartialEq,
        cosmwasm_schema::schemars::JsonSchema,
    )]
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[schemars(crate = "::cosmwasm_schema::schemars")]
    pub struct ChannelOpenInitOptions {
        /// The connection id on this chain.
        pub connection_id: String,
        /// The counterparty connection id on the counterparty chain.
        pub counterparty_connection_id: String,
        /// The counterparty port id. If not specified, [`crate::ibc::types::keys::HOST_PORT_ID`] is used.
        /// Currently, this contract only supports the host port.
        pub counterparty_port_id: Option<String>,
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
                .unwrap_or_else(|| crate::ibc::types::keys::HOST_PORT_ID.to_string())
        }
    }
}
