//! This module defines the state storage of the Contract.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, IbcChannel};
use cw_storage_plus::Item;

use super::{msg::options::ChannelOpenInitOptions, ContractError};

#[allow(clippy::module_name_repetitions)]
pub use channel::{State as ChannelState, Status as ChannelStatus};
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

mod channel {
    use cosmwasm_std::IbcOrder;

    use super::{cw_serde, IbcChannel};

    /// Status is the status of an IBC channel.
    #[cw_serde]
    pub enum Status {
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
    pub struct State {
        /// The IBC channel, as defined by cosmwasm.
        pub channel: IbcChannel,
        /// The status of the channel.
        pub channel_status: Status,
    }

    impl State {
        /// Creates a new [`ChannelState`]
        #[must_use]
        pub const fn new_open_channel(channel: IbcChannel) -> Self {
            Self {
                channel,
                channel_status: Status::Open,
            }
        }

        /// Checks if the channel is open
        #[must_use]
        pub const fn is_open(&self) -> bool {
            matches!(self.channel_status, Status::Open)
        }

        /// Closes the channel
        pub fn close(&mut self) {
            self.channel_status = Status::Closed;
        }

        /// Checks if the channel is [`IbcOrder::Ordered`]
        #[must_use]
        pub const fn is_ordered(&self) -> bool {
            matches!(self.channel.order, IbcOrder::Ordered)
        }
    }

    impl ToString for Status {
        fn to_string(&self) -> String {
            match self {
                Self::Uninitialized => "STATE_UNINITIALIZED_UNSPECIFIED".to_string(),
                Self::Init => "STATE_INIT".to_string(),
                Self::TryOpen => "STATE_TRYOPEN".to_string(),
                Self::Open => "STATE_OPEN".to_string(),
                Self::Closed => "STATE_CLOSED".to_string(),
                Self::Flushing => "STATE_FLUSHING".to_string(),
                Self::FlushComplete => "STATE_FLUSHCOMPLETE".to_string(),
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    use cw_ica_controller_v0_1_3::types::state as v0_1_3;
    use cw_ica_controller_v0_2_0::types::state as v0_2_0;
    use cw_ica_controller_v0_3_0::types::state as v0_3_0;
    use cw_ica_controller_v0_4_2 as v0_4_2;
    use cw_ica_controller_v0_5_0 as v0_5_0;

    mod v0_1_2 {
        use super::*;

        /// This is the contract state at version 0.1.2.
        #[cw_serde]
        pub struct ContractState {
            /// The address of the admin of the IBC application.
            pub admin: Addr,
            /// The Interchain Account (ICA) info needed to send packets.
            /// This is set during the handshake.
            #[serde(skip_serializing_if = "Option::is_none")]
            pub ica_info: Option<contract::IcaInfo>,
        }
    }

    #[test]
    fn test_migration_from_v0_1_2_to_v0_1_3() {
        let mock_state = v0_1_2::ContractState {
            admin: Addr::unchecked("admin"),
            ica_info: None,
        };

        let serialized = cosmwasm_std::to_json_binary(&mock_state).unwrap();

        let deserialized: v0_1_3::ContractState = cosmwasm_std::from_json(serialized).unwrap();

        let exp_state = v0_1_3::ContractState {
            admin: Addr::unchecked("admin"),
            ica_info: None,
            allow_channel_open_init: false,
        };

        assert_eq!(deserialized, exp_state);
    }

    #[test]
    fn test_migration_from_v0_1_3_to_v0_2_0() {
        let mock_state = v0_1_3::ContractState {
            admin: Addr::unchecked("admin"),
            ica_info: None,
            allow_channel_open_init: false,
        };

        let serialized = cosmwasm_std::to_json_binary(&mock_state).unwrap();

        let deserialized: v0_2_0::ContractState = cosmwasm_std::from_json(serialized).unwrap();

        let exp_state = v0_2_0::ContractState {
            admin: Addr::unchecked("admin"),
            ica_info: None,
            allow_channel_open_init: false,
            callback_address: None,
        };

        assert_eq!(deserialized, exp_state);
    }

    #[test]
    fn test_migration_from_v0_2_0_to_v0_3_0() {
        let mock_state = v0_2_0::ContractState {
            admin: Addr::unchecked("admin"),
            ica_info: None,
            allow_channel_open_init: false,
            callback_address: Some(Addr::unchecked("callback")),
        };

        let serialized = cosmwasm_std::to_json_binary(&mock_state).unwrap();

        let deserialized: v0_3_0::ContractState = cosmwasm_std::from_json(serialized).unwrap();

        let exp_state = v0_3_0::ContractState {
            ica_info: None,
            allow_channel_open_init: false,
            callback_address: Some(Addr::unchecked("callback")),
        };

        assert_eq!(deserialized, exp_state);
    }

    #[test]
    fn test_migration_from_v0_4_2_to_v0_5_0() {
        // Test channel open init options
        let mock_options = v0_4_2::types::msg::options::ChannelOpenInitOptions {
            connection_id: "connection-mock".to_string(),
            counterparty_connection_id: "counterparty-connection-mock".to_string(),
            counterparty_port_id: Some("counterparty-port-mock".to_string()),
            tx_encoding: Some(v0_4_2::ibc::types::metadata::TxEncoding::Protobuf),
        };

        let serialized = cosmwasm_std::to_json_binary(&mock_options).unwrap();

        let deserialized: v0_5_0::types::msg::options::ChannelOpenInitOptions =
            cosmwasm_std::from_json(serialized).unwrap();

        let exp_options = v0_5_0::types::msg::options::ChannelOpenInitOptions {
            connection_id: "connection-mock".to_string(),
            counterparty_connection_id: "counterparty-connection-mock".to_string(),
            counterparty_port_id: Some("counterparty-port-mock".to_string()),
            tx_encoding: Some(v0_5_0::ibc::types::metadata::TxEncoding::Protobuf),
            channel_ordering: None,
        };

        assert_eq!(deserialized, exp_options);

        // Test contract state
        let mock_state = v0_4_2::types::state::ContractState {
            ica_info: None,
            allow_channel_open_init: false,
            callback_address: Some(Addr::unchecked("callback")),
        };

        let serialized = cosmwasm_std::to_json_binary(&mock_state).unwrap();

        let deserialized: ContractState = cosmwasm_std::from_json(serialized).unwrap();

        let exp_state = ContractState {
            ica_info: None,
            callback_address: Some(Addr::unchecked("callback")),
        };

        assert_eq!(deserialized, exp_state);
    }

    #[test]
    fn test_migration_from_v0_5_0_to_v0_6_0() {
        let mock_channel_open_init_options = v0_5_0::types::msg::options::ChannelOpenInitOptions {
            connection_id: "connection-mock".to_string(),
            counterparty_connection_id: "counterparty-connection-mock".to_string(),
            counterparty_port_id: Some("counterparty-port-mock".to_string()),
            tx_encoding: Some(v0_5_0::ibc::types::metadata::TxEncoding::Proto3Json),
            channel_ordering: Some(cosmwasm_std::IbcOrder::Ordered),
        };

        let serialized = cosmwasm_std::to_json_binary(&mock_channel_open_init_options).unwrap();

        let deserialized: crate::types::msg::options::ChannelOpenInitOptions =
            cosmwasm_std::from_json(serialized).unwrap();

        let exp_channel_open_init_options = crate::types::msg::options::ChannelOpenInitOptions {
            connection_id: "connection-mock".to_string(),
            counterparty_connection_id: "counterparty-connection-mock".to_string(),
            counterparty_port_id: Some("counterparty-port-mock".to_string()),
            channel_ordering: Some(cosmwasm_std::IbcOrder::Ordered),
        };

        assert_eq!(deserialized, exp_channel_open_init_options);
    }
}
