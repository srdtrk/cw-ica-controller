//! # Packet
//!
//! This module contains the ICS-27 packet data and acknowledgement types.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{to_json_binary, CosmosMsg, Env, IbcMsg, IbcTimeout, StdError, StdResult};

pub use cosmos_sdk_proto::ibc::applications::interchain_accounts::v1::CosmosTx;
use cosmos_sdk_proto::traits::Message;

use crate::types::cosmos_msg::convert_to_proto_any;

use super::metadata::TxEncoding;

/// `DEFAULT_TIMEOUT_SECONDS` is the default timeout for [`IcaPacketData`]
pub const DEFAULT_TIMEOUT_SECONDS: u64 = 600;

/// `IcaPacketData` is comprised of a raw transaction, type of transaction and optional memo field.
/// Currently, the host only supports [protobuf](super::metadata::TxEncoding::Protobuf) or
/// [proto3json](super::metadata::TxEncoding::Proto3Json) serialized Cosmos transactions.
/// This contract only supports the protobuf encoding.
///
/// When protobuf is used, then the raw transaction must encoded using
/// [`CosmosTx`](cosmos_sdk_proto::ibc::applications::interchain_accounts::v1::CosmosTx).
#[allow(clippy::module_name_repetitions)]
#[cw_serde]
pub struct IcaPacketData {
    /// Type defines a classification of message issued from a controller
    /// chain to its associated interchain accounts host.
    ///
    /// There are two types of messages:
    /// * `0 (Unspecified)`: Default zero value enumeration. (Returns an error in host).
    /// * `1 (ExecuteTx)`: Execute a transaction on an interchain accounts host chain.
    ///
    /// `r#type` is used to avoid the reserved keyword `type`.
    #[serde(rename = "type")]
    pub r#type: u32,
    /// Data is the raw transaction data that will be sent to the interchain accounts host.
    /// Currently, the host only supports json (or proto) serialized Any messages.
    pub data: Vec<u8>,
    /// Memo is an optional field that can be used to attach a memo to a transaction.
    /// It is also caught by some ibc middleware to perform additional actions.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub memo: Option<String>,
}

impl IcaPacketData {
    /// Creates a new [`IcaPacketData`]
    #[must_use]
    pub fn new(data: Vec<u8>, memo: Option<String>) -> Self {
        Self {
            r#type: 1,
            data,
            memo,
        }
    }

    /// Creates a new [`IcaPacketData`] from a list of [`cosmos_sdk_proto::Any`] messages
    #[must_use]
    pub fn from_proto_anys(messages: Vec<cosmos_sdk_proto::Any>, memo: Option<String>) -> Self {
        let cosmos_tx = CosmosTx { messages };
        let data = cosmos_tx.encode_to_vec();
        Self::new(data, memo)
    }

    /// Creates a new [`IcaPacketData`] from a list of [`CosmosMsg`] messages
    ///
    /// # Errors
    ///
    /// Returns an error if the [`CosmosMsg`] cannot be serialized to [`cosmos_sdk_proto::Any`]
    /// when using the [`TxEncoding::Protobuf`] encoding.
    ///
    /// # Panics
    ///
    /// Panics if the [`CosmosMsg`] is not supported.
    ///
    /// The supported [`CosmosMsg`]s for [`TxEncoding::Protobuf`] are listed in [`convert_to_proto_any`].
    pub fn from_cosmos_msgs(
        messages: Vec<CosmosMsg>,
        encoding: &TxEncoding,
        memo: Option<String>,
        ica_address: &str,
    ) -> StdResult<Self> {
        match encoding {
            TxEncoding::Protobuf => {
                let proto_anys = messages.into_iter().try_fold(
                    vec![],
                    |mut acc, msg| -> StdResult<Vec<cosmos_sdk_proto::Any>> {
                        let proto_any = convert_to_proto_any(msg, ica_address.to_string())
                            .map_err(|e| StdError::generic_err(e.to_string()))?;
                        acc.push(proto_any);
                        Ok(acc)
                    },
                )?;
                Ok(Self::from_proto_anys(proto_anys, memo))
            }
            TxEncoding::Proto3Json => StdResult::Err(StdError::generic_err(
                "unsupported encoding: proto3json".to_string(),
            )),
        }
    }

    /// Creates an [`IbcMsg::SendPacket`] message from the [`IcaPacketData`]
    ///
    /// # Errors
    ///
    /// Returns an error if the [`IcaPacketData`] cannot be serialized to JSON.
    pub fn to_ibc_msg(
        &self,
        env: &Env,
        channel_id: impl Into<String>,
        timeout_seconds: Option<u64>,
    ) -> StdResult<IbcMsg> {
        let timeout_timestamp = env
            .block
            .time
            .plus_seconds(timeout_seconds.unwrap_or(DEFAULT_TIMEOUT_SECONDS));
        Ok(IbcMsg::SendPacket {
            channel_id: channel_id.into(),
            data: to_json_binary(&self)?,
            timeout: IbcTimeout::with_timestamp(timeout_timestamp),
        })
    }
}

/// contains the [`Data`] struct which is the acknowledgement to an ica packet
pub mod acknowledgement {
    use cosmwasm_std::Binary;

    use super::cw_serde;

    /// `Data` is the response to an ibc packet. It either contains a result or an error.
    #[cw_serde]
    pub enum Data {
        /// Result is the result of a successful transaction.
        Result(Binary),
        /// Error is the error message of a failed transaction.
        /// It is a string of the error message (not base64 encoded).
        Error(String),
    }
}

#[cfg(test)]
mod tests {
    use acknowledgement::Data as AcknowledgementData;
    use cosmwasm_std::{from_json, Binary};

    use super::*;

    #[test]
    fn test_acknowledgement() {
        // Test result:
        // The following bytes refer to `{"result":"c3VjY2Vzcw=="}`
        // where `c3VjY2Vzcw==` is the base64 encoding of `success`.
        let cw_success_binary = Binary(vec![
            123, 34, 114, 101, 115, 117, 108, 116, 34, 58, 34, 99, 51, 86, 106, 89, 50, 86, 122,
            99, 119, 61, 61, 34, 125,
        ]);
        let ack: AcknowledgementData = from_json(cw_success_binary).unwrap();
        assert_eq!(
            ack,
            AcknowledgementData::Result(Binary::from_base64("c3VjY2Vzcw==").unwrap())
        );

        // Test error:
        let error_bytes =
            br#"{"error":"ABCI code: 1: error handling packet: see events for details"}"#;
        let cw_error_binary = Binary(error_bytes.to_vec());
        let ack: AcknowledgementData = from_json(cw_error_binary).unwrap();
        assert_eq!(
            ack,
            AcknowledgementData::Error(
                "ABCI code: 1: error handling packet: see events for details".to_string()
            )
        );
    }
}
