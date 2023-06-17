use cosmwasm_std::{Addr, IbcChannel};
use cw_storage_plus::Item;
use serde::{Deserialize, Serialize};

/// STATE is the key used to store the state of the IBC application.
pub const STATE: Item<ContractState> = Item::new("state");

/// CHANNEL_STATE is the key used to store the state of the IBC application's channel.
pub const CHANNEL_STATE: Item<ContractChannelState> = Item::new("ica_channel");

/// ContractState is the state of the IBC application.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ContractState {
    pub admin: Addr,
    pub ica_address: Option<String>,
}

/// ContractChannelState is the state of the IBC application's channel.
/// This application only supports one channel.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ContractChannelState {
    pub channel: IbcChannel,
    pub channel_state: ChannelState,
}

/// ChannelState is the state of the IBC channel.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum ChannelState {
    Uninitialized,
    Init,
    TryOpen,
    Open,
    Closed,
}
