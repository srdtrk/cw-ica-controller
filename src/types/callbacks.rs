//! # Callbacks
//!
//! This module contains the callbacks message type that this contract can make to other
//! contracts upon channel and packet lifecycle events.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{
    to_json_binary, Addr, Binary, CosmosMsg, IbcChannel, IbcPacket, StdResult, WasmMsg,
};

use crate::ibc::types::{
    metadata::TxEncoding, packet::acknowledgement::Data as AcknowledgementData,
};

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
    /// serializes the message
    ///
    /// # Errors
    ///
    /// This function returns an error if the message cannot be serialized.
    pub fn into_json_binary(self) -> StdResult<Binary> {
        let msg = ReceiverExecuteMsg::ReceiveIcaCallback(self);
        to_json_binary(&msg)
    }

    /// `into_cosmos_msg` converts this message into a [`CosmosMsg`] message to be sent to
    /// the named contract.
    ///
    /// # Errors
    ///
    /// This function returns an error if the message cannot be serialized.
    pub fn into_cosmos_msg<C>(self, contract_addr: impl Into<String>) -> StdResult<CosmosMsg<C>>
    where
        C: Clone + std::fmt::Debug + PartialEq,
    {
        let execute = WasmMsg::Execute {
            contract_addr: contract_addr.into(),
            msg: self.into_json_binary()?,
            funds: vec![],
        };

        Ok(execute.into())
    }
}

/// This is just a helper to properly serialize the above message.
/// The actual receiver should include this variant in the larger ExecuteMsg enum
#[cw_serde]
enum ReceiverExecuteMsg {
    ReceiveIcaCallback(IcaControllerCallbackMsg),
}
