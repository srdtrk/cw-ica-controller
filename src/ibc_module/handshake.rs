#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    DepsMut, Env, IbcBasicResponse, IbcChannel, IbcChannelCloseMsg, IbcChannelConnectMsg,
    IbcChannelOpenMsg, IbcChannelOpenResponse, IbcOrder,
};

use super::types::{keys::HOST_PORT_ID, metadata::IcaMetadata};
use crate::{
    state::{ChannelState, STATE},
    ContractError,
};

/// Handles the `OpenInit` and `OpenTry` parts of the IBC handshake.
/// In this application, we only handle `OpenInit` messages.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_open(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelOpenMsg,
) -> Result<IbcChannelOpenResponse, ContractError> {
    match msg {
        IbcChannelOpenMsg::OpenInit { channel } => ibc_channel_open_init(deps, channel),
        IbcChannelOpenMsg::OpenTry { .. } => unimplemented!(),
    }
}

/// Handles the `OpenInit` part of the IBC handshake.
fn ibc_channel_open_init(
    deps: DepsMut,
    channel: IbcChannel,
) -> Result<IbcChannelOpenResponse, ContractError> {
    // Validate the channel ordering
    if channel.order != IbcOrder::Ordered {
        return Err(ContractError::InvalidChannelOrdering {});
    }
    // Validate the host port
    if channel.counterparty_endpoint.port_id != HOST_PORT_ID {
        return Err(ContractError::InvalidHostPort {});
    }

    // Deserialize the metadata
    let metadata: IcaMetadata = if channel.version.is_empty() {
        IcaMetadata::from_channel(&channel)
    } else {
        serde_json::from_str(&channel.version).map_err(|_| ContractError::UnknownDataType {})?
    };
    metadata.validate(&channel)?;

    if let Some(contract_state) = STATE.may_load(deps.storage)? {
        if contract_state.channel_state == ChannelState::Open {
            return Err(ContractError::ActiveChannelAlreadySet {});
        }
        let app_version = contract_state.channel.version;
        if !metadata.is_previous_version_equal(&app_version) {
            return Err(ContractError::InvalidVersion {
                expected: app_version,
                actual: metadata.version,
            });
        }
    }

    todo!()
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
