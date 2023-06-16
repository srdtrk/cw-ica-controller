use cosmwasm_std::IbcChannel;
use cw_storage_plus::Item;
use serde::{Deserialize, Serialize};

pub const STATE: Item<ContractState> = Item::new("ica_channel");

/// ContractState is the state of the IBC application
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ContractState {
    pub channel: IbcChannel,
    pub channel_state: ChannelState,
}

/// ChannelState is the state of the IBC channel
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum ChannelState {
    Uninitialized,
    Init,
    TryOpen,
    Open,
    Closed,
}
