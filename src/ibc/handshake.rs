//! This module contains the entry points for the IBC handshake.

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    DepsMut, Env, Ibc3ChannelOpenResponse, IbcBasicResponse, IbcChannel, IbcChannelCloseMsg,
    IbcChannelConnectMsg, IbcChannelOpenMsg, IbcChannelOpenResponse, IbcOrder,
};

use super::types::{keys::HOST_PORT_ID, metadata::IcaMetadata};
use crate::types::{
    state::{ChannelState, CHANNEL_STATE, STATE},
    ContractError,
};

/// Handles the `OpenInit` and `OpenTry` parts of the IBC handshake.
/// In this application, we only handle `OpenInit` messages since we are the ICA controller
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_open(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelOpenMsg,
) -> Result<IbcChannelOpenResponse, ContractError> {
    match msg {
        IbcChannelOpenMsg::OpenInit { channel } => ibc_channel_open::init(deps, channel),
        IbcChannelOpenMsg::OpenTry { .. } => unreachable!(),
    }
}

/// Handles the `OpenAck` and `OpenConfirm` parts of the IBC handshake.
/// In this application, we only handle `OpenAck` messages since we are the ICA controller
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_connect(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelConnectMsg,
) -> Result<IbcBasicResponse, ContractError> {
    match msg {
        IbcChannelConnectMsg::OpenAck {
            channel,
            counterparty_version,
        } => ibc_channel_open::on_acknowledgement(deps, channel, counterparty_version),
        IbcChannelConnectMsg::OpenConfirm { .. } => unreachable!(),
    }
}

/// Handles the `ChanCloseInit` and `ChanCloseConfirm` for the IBC module.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_close(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelCloseMsg,
) -> Result<IbcBasicResponse, ContractError> {
    match msg {
        IbcChannelCloseMsg::CloseInit { .. } => unimplemented!(),
        IbcChannelCloseMsg::CloseConfirm { channel } => ibc_channel_close::confirm(deps, channel),
    }
}

mod ibc_channel_open {
    use super::*;

    /// Handles the `OpenInit` part of the IBC handshake.
    pub fn init(
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

        // serde::Deserialize the metadata
        let metadata: IcaMetadata = if channel.version.is_empty() {
            // if empty, use create new metadata.
            IcaMetadata::from_channel(deps.as_ref(), &channel)
        } else {
            serde_json_wasm::from_str(&channel.version).map_err(|_| {
                ContractError::UnknownDataType(
                    "cannot unmarshal ICS-27 interchain accounts metadata".to_string(),
                )
            })?
        };
        metadata.validate(&channel)?;

        // Check if the channel is already exists
        if let Some(channel_state) = CHANNEL_STATE.may_load(deps.storage)? {
            // this contract can only store one active channel
            // if the channel is already open, return an error
            if channel_state.is_open() {
                return Err(ContractError::ActiveChannelAlreadySet {});
            }
            let app_version = channel_state.channel.version;
            if !metadata.is_previous_version_equal(&app_version) {
                return Err(ContractError::InvalidVersion {
                    expected: app_version,
                    actual: metadata.to_string(),
                });
            }
        }
        // Channel state need not be saved here, as it is tracked by wasmd during the handshake

        Ok(IbcChannelOpenResponse::Some(Ibc3ChannelOpenResponse {
            version: metadata.to_string(),
        }))
    }

    /// Handles the `OpenAck` part of the IBC handshake.
    pub fn on_acknowledgement(
        deps: DepsMut,
        channel: IbcChannel,
        counterparty_version: String,
    ) -> Result<IbcBasicResponse, ContractError> {
        // portID cannot be host chain portID
        // this is not possible since it is wasm.CONTRACT_ADDRESS
        // but we check it anyway since this is a recreation of the go code
        if channel.endpoint.port_id == HOST_PORT_ID {
            return Err(ContractError::InvalidControllerPort {});
        }

        // serde::Deserialize the metadata
        let metadata: IcaMetadata =
            serde_json_wasm::from_str(&counterparty_version).map_err(|_| {
                ContractError::UnknownDataType(
                    "cannot unmarshal ICS-27 interchain accounts metadata".to_string(),
                )
            })?;
        metadata.validate(&channel)?;

        // Check if the address is empty
        if metadata.address.is_empty() {
            return Err(ContractError::InvalidAddress {});
        }
        // save the address to the contract state
        STATE.update(
            deps.storage,
            |mut contract_state| -> Result<_, ContractError> {
                contract_state.set_ica_info(
                    metadata.address,
                    &channel.endpoint.channel_id,
                    metadata.encoding,
                );
                Ok(contract_state)
            },
        )?;

        // Save the channel state
        CHANNEL_STATE.save(deps.storage, &ChannelState::new_open_channel(channel))?;

        // Return the response, emit events if needed. Core IBC modules will emit the events regardless.
        Ok(IbcBasicResponse::default())
    }
}

mod ibc_channel_close {
    use super::*;
    /// Handles the `ChanCloseConfirm` for the IBC module.
    pub fn confirm(deps: DepsMut, channel: IbcChannel) -> Result<IbcBasicResponse, ContractError> {
        // Validate that this is the stored channel
        let mut channel_state = CHANNEL_STATE.load(deps.storage)?;
        if channel_state.channel != channel {
            return Err(ContractError::InvalidChannelInContractState {});
        }

        // Update the channel state
        channel_state.close();
        CHANNEL_STATE.save(deps.storage, &channel_state)?;

        // Return the response, emit events if needed
        Ok(IbcBasicResponse::default())
    }
}
