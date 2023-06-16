use cosmwasm_std::IbcChannel;
use cw_storage_plus::Item;

const ICA_CHANNEL: Item<IbcChannel> = Item::new("ica_channel");
