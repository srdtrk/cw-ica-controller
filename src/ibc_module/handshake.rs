#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    DepsMut, Env, IbcBasicResponse, IbcChannelCloseMsg, IbcChannelConnectMsg, IbcChannelOpenMsg,
    IbcChannelOpenResponse,
};

use crate::ContractError;

/// Handles the `OpenInit` and `OpenTry` parts of the IBC handshake.
/// In this application, we only handle `OpenInit` messages.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_open(
    _deps: DepsMut,
    _env: Env,
    msg: IbcChannelOpenMsg,
) -> Result<IbcChannelOpenResponse, ContractError> {
    match msg {
        IbcChannelOpenMsg::OpenInit { channel } => todo!(),
        IbcChannelOpenMsg::OpenTry { .. } => unimplemented!(),
    }
}

/// Handles the `OpenAck` and `OpenConfirm` parts of the IBC handshake.
/// In this application, we only handle `OpenAck` messages.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_connect(
    _deps: DepsMut,
    _env: Env,
    msg: IbcChannelConnectMsg,
) -> Result<IbcBasicResponse, ContractError> {
    match msg {
        IbcChannelConnectMsg::OpenAck {
            channel,
            counterparty_version,
        } => todo!(),
        IbcChannelConnectMsg::OpenConfirm { .. } => unimplemented!(),
    }
}

/// Handles the `OnChanCloseConfirm` for the IBC module.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_close(
    _deps: DepsMut,
    _env: Env,
    msg: IbcChannelCloseMsg,
) -> Result<IbcBasicResponse, ContractError> {
    match msg {
        IbcChannelCloseMsg::CloseInit { .. } => unimplemented!(),
        IbcChannelCloseMsg::CloseConfirm { channel } => todo!(),
    }
}
