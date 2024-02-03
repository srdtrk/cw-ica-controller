//! This module contains the entry points for the IBC handshake.

use cosmwasm_std::entry_point;
use cosmwasm_std::{
    DepsMut, Env, Ibc3ChannelOpenResponse, IbcBasicResponse, IbcChannel, IbcChannelCloseMsg,
    IbcChannelConnectMsg, IbcChannelOpenMsg, IbcChannelOpenResponse,
};

use super::types::{keys, metadata::IcaMetadata};
use crate::types::{
    state::{self, ChannelState},
    ContractError,
};

/// Handles the `OpenInit` and `OpenTry` parts of the IBC handshake.
/// In this application, we only handle `OpenInit` messages since we are the ICA controller
///
/// # Errors
///
/// This function returns an error if:
///
/// - The channel is already open.
/// - `allow_channel_open_init` is disabled.
/// - The host port is invalid.
/// - Version metadata is invalid.
#[entry_point]
#[allow(clippy::needless_pass_by_value)] // entry point needs this signature
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
///
/// # Errors
///
/// This function returns an error if:
///
/// - The host port is invalid.
/// - The version metadata is invalid.
/// - The ICA address is empty.
#[entry_point]
#[allow(clippy::needless_pass_by_value)] // entry point needs this signature
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
///
/// # Errors
///
/// This function returns an error if:
///
/// - The channel is not stored in the contract state.
/// - The channel is not open.
/// - The channel is not the same as the stored channel.
#[entry_point]
#[allow(clippy::needless_pass_by_value)] // entry point needs this signature
pub fn ibc_channel_close(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelCloseMsg,
) -> Result<IbcBasicResponse, ContractError> {
    match msg {
        IbcChannelCloseMsg::CloseInit { channel } => ibc_channel_close::init(deps, channel),
        IbcChannelCloseMsg::CloseConfirm { channel } => ibc_channel_close::confirm(deps, channel),
    }
}

mod ibc_channel_open {
    use crate::types::callbacks::IcaControllerCallbackMsg;

    use super::{
        keys, state, ChannelState, ContractError, DepsMut, Ibc3ChannelOpenResponse,
        IbcBasicResponse, IbcChannel, IbcChannelOpenResponse, IcaMetadata,
    };

    /// Handles the `OpenInit` part of the IBC handshake.
    #[allow(clippy::needless_pass_by_value)]
    pub fn init(
        deps: DepsMut,
        channel: IbcChannel,
    ) -> Result<IbcChannelOpenResponse, ContractError> {
        if !state::ALLOW_CHANNEL_OPEN_INIT
            .load(deps.storage)
            .unwrap_or_default()
        {
            return Err(ContractError::ChannelOpenInitNotAllowed);
        }

        state::ALLOW_CHANNEL_OPEN_INIT.save(deps.storage, &false)?;

        // Validate the host port
        if channel.counterparty_endpoint.port_id != keys::HOST_PORT_ID {
            return Err(ContractError::InvalidHostPort);
        }

        // serde::Deserialize the metadata
        let metadata: IcaMetadata = if channel.version.is_empty() {
            // if empty, use create new metadata.
            IcaMetadata::from_channel(deps.as_ref(), &channel)?
        } else {
            serde_json_wasm::from_str(&channel.version).map_err(|_| {
                ContractError::UnknownDataType(
                    "cannot unmarshal ICS-27 interchain accounts metadata".to_string(),
                )
            })?
        };
        metadata.validate(&channel)?;

        // Check if the channel is already exists
        if let Some(channel_state) = state::CHANNEL_STATE.may_load(deps.storage)? {
            // this contract can only store one active channel
            // if the channel is already open, return an error
            if channel_state.is_open() {
                return Err(ContractError::ActiveChannelAlreadySet);
            }
        }
        // Channel state need not be saved here, as it is tracked by wasmd during the handshake

        Ok(IbcChannelOpenResponse::Some(Ibc3ChannelOpenResponse {
            version: metadata.to_string(),
        }))
    }

    /// Handles the `OpenAck` part of the IBC handshake.
    #[allow(clippy::needless_pass_by_value)]
    pub fn on_acknowledgement(
        deps: DepsMut,
        mut channel: IbcChannel,
        counterparty_version: String,
    ) -> Result<IbcBasicResponse, ContractError> {
        let mut state = state::STATE.load(deps.storage)?;

        // portID cannot be host chain portID
        // this is not possible since it is wasm.CONTRACT_ADDRESS
        // but we check it anyway since this is a recreation of the go code
        if channel.endpoint.port_id == keys::HOST_PORT_ID {
            return Err(ContractError::InvalidControllerPort);
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
            return Err(ContractError::InvalidIcaAddress);
        }

        // update state with the ica info
        state.set_ica_info(
            &metadata.address,
            &channel.endpoint.channel_id,
            metadata.encoding.clone(),
        );
        state::STATE.save(deps.storage, &state)?;

        channel.version = counterparty_version;

        // Save the channel state
        state::CHANNEL_STATE.save(
            deps.storage,
            &ChannelState::new_open_channel(channel.clone()),
        )?;

        // make callback if needed
        if let Some(callback_address) = state.callback_address {
            let callback_msg = IcaControllerCallbackMsg::OnChannelOpenAckCallback {
                channel,
                ica_address: metadata.address,
                tx_encoding: metadata.encoding,
            }
            .into_cosmos_msg(callback_address)?;

            Ok(IbcBasicResponse::default().add_message(callback_msg))
        } else {
            Ok(IbcBasicResponse::default())
        }
    }
}

mod ibc_channel_close {
    use super::{state, ContractError, DepsMut, IbcBasicResponse, IbcChannel};

    /// Handles the `ChanClosedInit` for the IBC module.
    #[allow(clippy::needless_pass_by_value)]
    pub fn init(deps: DepsMut, channel: IbcChannel) -> Result<IbcBasicResponse, ContractError> {
        if !state::ALLOW_CHANNEL_CLOSE_INIT
            .load(deps.storage)
            .unwrap_or_default()
        {
            return Err(ContractError::ChannelCloseInitNotAllowed);
        };

        state::ALLOW_CHANNEL_CLOSE_INIT.save(deps.storage, &false)?;

        // Validate that this is the stored channel
        let mut channel_state = state::CHANNEL_STATE.load(deps.storage)?;
        if channel_state.channel != channel {
            return Err(ContractError::InvalidChannelInContractState);
        }
        if !channel_state.is_open() {
            return Err(ContractError::InvalidChannelStatus {
                expected: state::ChannelStatus::Open.to_string(),
                actual: channel_state.channel_status.to_string(),
            });
        }

        // Update the channel state
        channel_state.close();
        state::CHANNEL_STATE.save(deps.storage, &channel_state)?;

        // Return the response, emit events if needed
        Ok(IbcBasicResponse::default())
    }

    /// Handles the `ChanCloseConfirm` for the IBC module.
    #[allow(clippy::needless_pass_by_value)]
    pub fn confirm(deps: DepsMut, channel: IbcChannel) -> Result<IbcBasicResponse, ContractError> {
        // Validate that this is the stored channel
        let mut channel_state = state::CHANNEL_STATE.load(deps.storage)?;
        if channel_state.channel != channel {
            return Err(ContractError::InvalidChannelInContractState);
        }
        if !channel_state.is_open() {
            return Err(ContractError::InvalidChannelStatus {
                expected: state::ChannelStatus::Open.to_string(),
                actual: channel_state.channel_status.to_string(),
            });
        }

        // Update the channel state
        channel_state.close();
        state::CHANNEL_STATE.save(deps.storage, &channel_state)?;

        // Return the response, emit events if needed
        Ok(IbcBasicResponse::default())
    }
}
