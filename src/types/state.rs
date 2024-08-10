//! This module defines the state storage of the Contract.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, IbcChannel};
use cw_storage_plus::Item;

use super::{msg::options::ChannelOpenInitOptions, ContractError};

#[allow(clippy::module_name_repetitions)]
pub use channel::{ChannelState, ChannelStatus};
#[allow(clippy::module_name_repetitions)]
pub use contract::State as ContractState;

/// The item used to store the state of the IBC application.
pub const STATE: Item<ContractState> = Item::new("state");

/// The item used to store the state of the IBC application's channel.
pub const CHANNEL_STATE: Item<ChannelState> = Item::new("ica_channel");

/// The item used to store the channel open init options.
pub const CHANNEL_OPEN_INIT_OPTIONS: Item<ChannelOpenInitOptions> =
    Item::new("channel_open_init_options");

/// The item used to store whether or not channel open init is allowed.
/// Used to prevent relayers from opening channels. This right is reserved to the contract.
pub const ALLOW_CHANNEL_OPEN_INIT: Item<bool> = Item::new("allow_channel_open_init");

/// The item used to store whether or not channel close init is allowed.
/// Used to prevent relayers from closing channels. This right is reserved to the contract.
pub const ALLOW_CHANNEL_CLOSE_INIT: Item<bool> = Item::new("allow_channel_close_init");

/// The item used to store the paths of an ICA query until its `SendPacket` response is received.
/// Once the response is received, it is moved to the [`PENDING_QUERIES`] map and deleted from this item.
/// This is used to ensure that the correct sequence is recorded for the response.
#[cfg(feature = "query")]
pub const QUERY: Item<Vec<(String, bool)>> = Item::new("pending_query");

/// `PENDING_QUERIES` is the map of pending queries.
/// It maps `channel_id`, and sequence to the query path.
#[cfg(feature = "query")]
pub const PENDING_QUERIES: cw_storage_plus::Map<(&str, u64), Vec<(String, bool)>> =
    cw_storage_plus::Map::new("pending_queries");

mod contract {
    use crate::ibc::types::metadata::TxEncoding;

    use cosmwasm_schema::schemars::JsonSchema;

    use super::{cw_serde, Addr, ContractError};

    /// State is the state of the contract.
    #[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq, JsonSchema)]
    #[allow(clippy::derive_partial_eq_without_eq)]
    pub struct State {
        /// The Interchain Account (ICA) info needed to send packets.
        /// This is set during the handshake.
        #[serde(default)]
        pub ica_info: Option<IcaInfo>,
        /// The address of the callback contract.
        #[serde(default)]
        pub callback_address: Option<Addr>,
    }

    impl State {
        /// Creates a new [`State`]
        #[must_use]
        pub const fn new(callback_address: Option<Addr>) -> Self {
            Self {
                ica_info: None,
                // We always allow the first `MsgChannelOpenInit` message.
                callback_address,
            }
        }

        /// Gets the ICA info
        ///
        /// # Errors
        ///
        /// Returns an error if the ICA info is not set.
        pub fn get_ica_info(&self) -> Result<IcaInfo, ContractError> {
            self.ica_info
                .as_ref()
                .map_or(Err(ContractError::IcaInfoNotSet), |s| Ok(s.clone()))
        }

        /// Sets the ICA info
        pub fn set_ica_info(
            &mut self,
            ica_address: impl Into<String>,
            channel_id: impl Into<String>,
            encoding: TxEncoding,
        ) {
            self.ica_info = Some(IcaInfo::new(ica_address, channel_id, encoding));
        }

        /// Deletes the ICA info
        pub fn delete_ica_info(&mut self) {
            self.ica_info = None;
        }
    }

    /// IcaInfo is the ICA address and channel ID.
    #[cw_serde]
    pub struct IcaInfo {
        pub ica_address: String,
        pub channel_id: String,
        pub encoding: TxEncoding,
    }

    impl IcaInfo {
        /// Creates a new [`IcaInfo`]
        pub fn new(
            ica_address: impl Into<String>,
            channel_id: impl Into<String>,
            encoding: TxEncoding,
        ) -> Self {
            Self {
                ica_address: ica_address.into(),
                channel_id: channel_id.into(),
                encoding,
            }
        }
    }
}

#[allow(clippy::module_name_repetitions)]
mod channel {
    use cosmwasm_std::IbcOrder;

    use super::{cw_serde, IbcChannel};

    /// Status is the status of an IBC channel.
    #[cw_serde]
    pub enum ChannelStatus {
        /// Uninitialized is the default state of the channel.
        #[serde(rename = "STATE_UNINITIALIZED_UNSPECIFIED")]
        Uninitialized,
        /// Init is the state of the channel when it is created.
        #[serde(rename = "STATE_INIT")]
        Init,
        /// TryOpen is the state of the channel when it is trying to open.
        #[serde(rename = "STATE_TRYOPEN")]
        TryOpen,
        /// Open is the state of the channel when it is open.
        #[serde(rename = "STATE_OPEN")]
        Open,
        /// Closed is the state of the channel when it is closed.
        #[serde(rename = "STATE_CLOSED")]
        Closed,
        /// The channel has just accepted the upgrade handshake attempt and
        /// is flushing in-flight packets. Added in `ibc-go` v8.1.0.
        #[serde(rename = "STATE_FLUSHING")]
        Flushing,
        /// The channel has just completed flushing any in-flight packets.
        /// Added in `ibc-go` v8.1.0.
        #[serde(rename = "STATE_FLUSHCOMPLETE")]
        FlushComplete,
    }

    /// State is the state of the IBC application's channel.
    /// This application only supports one channel.
    #[cw_serde]
    pub struct ChannelState {
        /// The IBC channel, as defined by cosmwasm.
        pub channel: IbcChannel,
        /// The status of the channel.
        pub channel_status: ChannelStatus,
    }

    impl ChannelState {
        /// Creates a new [`ChannelState`]
        #[must_use]
        pub const fn new_open_channel(channel: IbcChannel) -> Self {
            Self {
                channel,
                channel_status: ChannelStatus::Open,
            }
        }

        /// Checks if the channel is open
        #[must_use]
        pub const fn is_open(&self) -> bool {
            matches!(self.channel_status, ChannelStatus::Open)
        }

        /// Closes the channel
        pub fn close(&mut self) {
            self.channel_status = ChannelStatus::Closed;
        }

        /// Checks if the channel is [`IbcOrder::Ordered`]
        #[must_use]
        pub const fn is_ordered(&self) -> bool {
            matches!(self.channel.order, IbcOrder::Ordered)
        }
    }

    impl std::fmt::Display for ChannelStatus {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            match self {
                Self::Uninitialized => write!(f, "STATE_UNINITIALIZED_UNSPECIFIED"),
                Self::Init => write!(f, "STATE_INIT"),
                Self::TryOpen => write!(f, "STATE_TRYOPEN"),
                Self::Open => write!(f, "STATE_OPEN"),
                Self::Closed => write!(f, "STATE_CLOSED"),
                Self::Flushing => write!(f, "STATE_FLUSHING"),
                Self::FlushComplete => write!(f, "STATE_FLUSHCOMPLETE"),
            }
        }
    }
}

/// This module defines the types stored in the state for ICA queries.
#[cfg(feature = "query")]
pub mod ica_query {
    use super::cw_serde;

    /// PendingQuery is the query packet that is pending a response.
    #[cw_serde]
    pub struct PendingQuery {
        /// The source channel ID of the query packet.
        pub channel_id: String,
        /// The sequence number of the query packet.
        pub sequence: u64,
        /// The gRPC query path.
        pub path: String,
        /// Whether the query was [`cosmwasm_std::QueryRequest::Stargate`] or not.
        pub is_stargate: bool,
    }
}
