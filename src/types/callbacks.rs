//! # Callbacks
//!
//! This module contains the callbacks that this contract can make to other contracts upon
//! channel and packet lifecycle events.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{to_binary, Addr, CosmosMsg, IbcChannel, IbcPacket, StdResult, WasmMsg};

use crate::ibc::types::{metadata::TxEncoding, packet::acknowledgement::AcknowledgementData};

/// IcaControllerCallbackMsg is the type of message that this contract can send to other contracts.
#[cw_serde]
pub enum IcaControllerCallbackMsg {
    /// OnAcknowledgementPacketCallback is the callback that this contract makes to other contracts
    /// when it receives an acknowledgement packet.
    OnAcknowledgementPacketCallback {
        /// The deserialized ICA acknowledgement data
        ica_acknowledgement: AcknowledgementData,
        /// The original packet that was sent
        original_packet: IbcPacket,
        /// The relayer that submitted acknowledgement packet
        relayer: Addr,
    },
    /// OnTimeoutPacketCallback is the callback that this contract makes to other contracts
    /// when it receives a timeout packet.
    OnTimeoutPacketCallback {
        /// The original packet that was sent
        original_packet: IbcPacket,
        /// The relayer that submitted acknowledgement packet
        relayer: Addr,
    },
    /// OnChannelOpenAckCallback is the callback that this contract makes to other contracts
    /// when it receives a channel open acknowledgement.
    OnChannelOpenAckCallback {
        /// The channel that was opened.
        channel: IbcChannel,
        /// The address of the interchain account that was created.
        ica_address: String,
        /// The tx encoding this ICA channel uses.
        tx_encoding: TxEncoding,
    },
}

impl IcaControllerCallbackMsg {
    /// into_cosmos_msg converts this message into a WasmMsg::Execute message to be sent to the
    /// named contract.
    pub fn into_cosmos_msg(&self, contract_addr: impl Into<String>) -> StdResult<CosmosMsg> {
        let execute = WasmMsg::Execute {
            contract_addr: contract_addr.into(),
            msg: to_binary(&self)?,
            funds: vec![],
        };

        Ok(execute.into())
    }
}
