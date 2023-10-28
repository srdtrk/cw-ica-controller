//! This module defines the state storage of the Contract.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, IbcChannel};
use cw_storage_plus::Item;

use super::ContractError;

pub use channel::ChannelState;
pub use contract::{CallbackCounter, ContractState};

/// The item used to store the state of the IBC application.
pub const STATE: Item<ContractState> = Item::new("state");

/// The item used to store the state of the IBC application's channel.
pub const CHANNEL_STATE: Item<ChannelState> = Item::new("ica_channel");

/// The item used to store the successful and erroneous callbacks in store.
pub const CALLBACK_COUNTER: Item<CallbackCounter> = Item::new("callback_counter");

mod contract {
    use crate::ibc::types::metadata::TxEncoding;

    use super::*;

    /// ContractState is the state of the IBC application.
    #[cw_serde]
    pub struct ContractState {
        /// The address of the admin of the IBC application.
        pub admin: Addr,
        /// The Interchain Account (ICA) info needed to send packets.
        /// This is set during the handshake.
        #[serde(skip_serializing_if = "Option::is_none")]
        pub ica_info: Option<IcaInfo>,
        /// If true, the IBC application will accept `MsgChannelOpenInit` messages.
        #[serde(default)]
        pub allow_channel_open_init: bool,
    }

    impl ContractState {
        /// Creates a new ContractState
        pub fn new(admin: Addr) -> Self {
            Self {
                admin,
                ica_info: None,
                // We always allow the first `MsgChannelOpenInit` message.
                allow_channel_open_init: true,
            }
        }

        /// Checks if the address is the admin
        pub fn verify_admin(&self, sender: impl Into<String>) -> Result<(), ContractError> {
            if self.admin == sender.into() {
                Ok(())
            } else {
                Err(ContractError::Unauthorized {})
            }
        }

        /// Checks if channel open init is allowed
        pub fn verify_open_init_allowed(&self) -> Result<(), ContractError> {
            if self.allow_channel_open_init {
                Ok(())
            } else {
                Err(ContractError::ChannelOpenInitNotAllowed {})
            }
        }

        /// Gets the ICA info
        pub fn get_ica_info(&self) -> Result<IcaInfo, ContractError> {
            if let Some(ica_info) = &self.ica_info {
                Ok(ica_info.clone())
            } else {
                Err(ContractError::IcaInfoNotSet {})
            }
        }

        /// Disables channel open init
        pub fn disable_channel_open_init(&mut self) {
            self.allow_channel_open_init = false;
        }

        /// Enables channel open init
        pub fn enable_channel_open_init(&mut self) {
            self.allow_channel_open_init = true;
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

    /// CallbackCounter tracks the number of callbacks in store.
    #[cw_serde]
    #[derive(Default)]
    pub struct CallbackCounter {
        /// The number of successful callbacks.
        pub success: u32,
        /// The number of erroneous callbacks.
        pub error: u32,
        /// The number of timeout callbacks.
        /// The channel is closed after a timeout due to the semantics of ordered channels.
        pub timeout: u32,
    }

    impl IcaInfo {
        /// Creates a new IcaInfo
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

    impl CallbackCounter {
        /// Increments the success counter
        pub fn success(&mut self) {
            self.success += 1;
        }

        /// Increments the error counter
        pub fn error(&mut self) {
            self.error += 1;
        }

        /// Increments the timeout counter
        pub fn timeout(&mut self) {
            self.timeout += 1;
        }
    }
}

mod channel {
    use super::*;

    /// ChannelState is the state of the IBC channel.
    #[cw_serde]
    pub enum ChannelStatus {
        #[serde(rename = "STATE_UNINITIALIZED_UNSPECIFIED")]
        Uninitialized,
        #[serde(rename = "STATE_INIT")]
        Init,
        #[serde(rename = "STATE_TRYOPEN")]
        TryOpen,
        #[serde(rename = "STATE_OPEN")]
        Open,
        #[serde(rename = "STATE_CLOSED")]
        Closed,
    }

    /// ContractChannelState is the state of the IBC application's channel.
    /// This application only supports one channel.
    #[cw_serde]
    pub struct ChannelState {
        /// The IBC channel, as defined by cosmwasm.
        pub channel: IbcChannel,
        /// The status of the channel.
        pub channel_status: ChannelStatus,
    }

    impl ChannelState {
        /// Creates a new ChannelState
        pub fn new_open_channel(channel: IbcChannel) -> Self {
            Self {
                channel,
                channel_status: ChannelStatus::Open,
            }
        }

        /// Checks if the channel is open
        pub fn is_open(&self) -> bool {
            self.channel_status == ChannelStatus::Open
        }

        /// Closes the channel
        pub fn close(&mut self) {
            self.channel_status = ChannelStatus::Closed;
        }
    }
}
