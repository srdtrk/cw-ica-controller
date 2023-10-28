//! This module contains the entry points for:
//! - The IBC packet acknowledgement.
//! - The IBC packet timeout.
//! - The IBC packet receive.

use cosmwasm_std::entry_point;
use cosmwasm_std::{
    from_binary, DepsMut, Env, IbcBasicResponse, IbcPacketAckMsg, IbcPacketReceiveMsg,
    IbcPacketTimeoutMsg, IbcReceiveResponse, Never,
};

use crate::types::{
    state::{CALLBACK_COUNTER, CHANNEL_STATE},
    ContractError,
};

use super::types::{events, packet::acknowledgement::AcknowledgementData};

/// Implements the IBC module's `OnAcknowledgementPacket` handler.
#[entry_point]
pub fn ibc_packet_ack(
    deps: DepsMut,
    _env: Env,
    ack: IbcPacketAckMsg,
) -> Result<IbcBasicResponse, ContractError> {
    // This lets the ICA controller know whether or not the sent transactions succeeded.
    match from_binary(&ack.acknowledgement.data)? {
        AcknowledgementData::Result(res) => ibc_packet_ack::success(deps, ack.original_packet, res),
        AcknowledgementData::Error(err) => ibc_packet_ack::error(deps, ack.original_packet, err),
    }
}

/// Handles the `PacketTimeout` for the IBC module.
#[entry_point]
pub fn ibc_packet_timeout(
    deps: DepsMut,
    _env: Env,
    _msg: IbcPacketTimeoutMsg,
) -> Result<IbcBasicResponse, ContractError> {
    // Increment the callback counter.
    CALLBACK_COUNTER.update(deps.storage, |mut cc| -> Result<_, ContractError> {
        cc.timeout();
        Ok(cc)
    })?;
    // Due to the semantics of ordered channels, the underlying channel end is closed.
    CHANNEL_STATE.update(
        deps.storage,
        |mut channel_state| -> Result<_, ContractError> {
            channel_state.close();
            Ok(channel_state)
        },
    )?;

    Ok(IbcBasicResponse::default())
}

/// Handles the `PacketReceive` for the IBC module.
#[entry_point]
pub fn ibc_packet_receive(
    _deps: DepsMut,
    _env: Env,
    _msg: IbcPacketReceiveMsg,
) -> Result<IbcReceiveResponse, Never> {
    // An ICA controller cannot receive packets, so this is a no-op.
    // It must be implemented to satisfy the wasmd interface.
    unreachable!("ICA controller cannot receive packets")
}

mod ibc_packet_ack {
    use cosmwasm_std::{Binary, IbcPacket};

    use crate::types::state::CALLBACK_COUNTER;

    use super::*;

    /// Handles the successful acknowledgement of an ica packet. This means that the
    /// transaction was successfully executed on the host chain.
    pub fn success(
        deps: DepsMut,
        packet: IbcPacket,
        res: Binary,
    ) -> Result<IbcBasicResponse, ContractError> {
        // Handle the success case.
        CALLBACK_COUNTER.update(deps.storage, |mut counter| -> Result<_, ContractError> {
            counter.success();
            Ok(counter)
        })?;
        Ok(IbcBasicResponse::default().add_event(events::packet_ack::success(&packet, &res)))
    }

    /// Handles the unsuccessful acknowledgement of an ica packet. This means that the
    /// transaction failed to execute on the host chain.
    pub fn error(
        deps: DepsMut,
        packet: IbcPacket,
        err: String,
    ) -> Result<IbcBasicResponse, ContractError> {
        // Handle the error.
        CALLBACK_COUNTER.update(deps.storage, |mut counter| -> Result<_, ContractError> {
            counter.error();
            Ok(counter)
        })?;
        Ok(IbcBasicResponse::default().add_event(events::packet_ack::error(&packet, &err)))
    }
}
