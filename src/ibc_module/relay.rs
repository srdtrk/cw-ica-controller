#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{DepsMut, Env, IbcBasicResponse, IbcPacketTimeoutMsg};

use crate::{
    state::{ChannelState, STATE},
    ContractError,
};

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_packet_timeout(
    deps: DepsMut,
    _env: Env,
    _msg: IbcPacketTimeoutMsg,
) -> Result<IbcBasicResponse, ContractError> {
    // Due to the semantics of ordered channels, the underlying channel end is closed.
    STATE.update(
        deps.storage,
        |mut contract_state| -> Result<_, ContractError> {
            contract_state.channel_state = ChannelState::Closed;
            Ok(contract_state)
        },
    )?;

    Ok(IbcBasicResponse::default())
}
