use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, IbcChannel};
use cw_storage_plus::Item;

pub use channel::ChannelState;
pub use contract::ContractState;

/// STATE is the item used to store the state of the IBC application.
pub const STATE: Item<ContractState> = Item::new("state");

/// CHANNEL_STATE is the item used to store the state of the IBC application's channel.
pub const CHANNEL_STATE: Item<ChannelState> = Item::new("ica_channel");

mod contract {
    use crate::types::ContractError;

    use super::*;

    /// ContractState is the state of the IBC application.
    #[cw_serde]
    pub struct ContractState {
        pub admin: Addr,
        pub ica_address: Option<String>,
    }

    impl ContractState {
        /// Creates a new ContractState
        pub fn new(admin: Addr, ica_address: Option<String>) -> Self {
            Self { admin, ica_address }
        }

        /// Checks if the address is the admin
        pub fn is_admin(&self, sender: impl Into<String>) -> bool {
            self.admin == sender.into()
        }

        /// Gets the ICA address
        pub fn get_ica_address(&self) -> Result<String, ContractError> {
            if let Some(ica_address) = &self.ica_address {
                Ok(ica_address.clone())
            } else {
                Err(ContractError::IcaAddressNotSet {})
            }
        }

        /// Sets the ICA address
        pub fn set_ica_address(&mut self, ica_address: String) {
            self.ica_address = Some(ica_address);
        }
    }
}

mod channel {
    use super::*;

    /// ChannelState is the state of the IBC channel.
    #[cw_serde]
    pub enum ChannelStatus {
        Uninitialized,
        Init,
        TryOpen,
        Open,
        Closed,
    }

    /// ContractChannelState is the state of the IBC application's channel.
    /// This application only supports one channel.
    #[cw_serde]
    pub struct ChannelState {
        pub channel: IbcChannel,
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
