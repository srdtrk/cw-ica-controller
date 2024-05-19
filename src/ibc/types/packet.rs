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
    #[cfg(feature = "query")]
    pub fn from_cosmos_msgs(
        #[cfg(feature = "export")] storage: &mut dyn cosmwasm_std::Storage,
        messages: Vec<CosmosMsg>,
        queries: Option<Vec<cosmwasm_std::QueryRequest<cosmwasm_std::Empty>>>,
        encoding: &TxEncoding,
        memo: Option<String>,
        ica_address: &str,
    ) -> StdResult<Self> {
        match encoding {
            TxEncoding::Protobuf => {
                use crate::types::query_msg;

                let mut proto_anys = messages
                    .into_iter()
                    .map(|msg| -> StdResult<cosmos_sdk_proto::Any> {
                        convert_to_proto_any(msg, ica_address.to_string())
                            .map_err(|e| StdError::generic_err(e.to_string()))
                    })
                    .collect::<StdResult<Vec<cosmos_sdk_proto::Any>>>()?;

                if let Some(queries) = queries {
                    if !queries.is_empty() {
                        let (abci_queries, _paths): (
                            Vec<query_msg::proto::AbciQueryRequest>,
                            Vec<(String, bool)>,
                        ) = queries.into_iter().fold((vec![], vec![]), |mut acc, msg| {
                            let (path, data, is_stargate) = query_msg::query_to_protobuf(msg);

                            acc.1.push((path.clone(), is_stargate));
                            acc.0
                                .push(query_msg::proto::AbciQueryRequest { path, data });

                            acc
                        });

                        #[cfg(feature = "export")]
                        #[allow(clippy::used_underscore_binding)]
                        crate::types::state::QUERY.save(storage, &_paths)?;

                        let query_msg = query_msg::proto::MsgModuleQuerySafe {
                            signer: ica_address.to_string(),
                            requests: abci_queries,
                        };

                        proto_anys.push(cosmos_sdk_proto::Any::from_msg(&query_msg).map_err(
                            |e| StdError::generic_err(format!("failed to convert query msg: {e}")),
                        )?);
                    }
                }

                Ok(Self::from_proto_anys(proto_anys, memo))
            }
            TxEncoding::Proto3Json => StdResult::Err(StdError::generic_err(
                "unsupported encoding: proto3json".to_string(),
            )),
        }
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
    #[cfg(not(feature = "query"))]
    pub fn from_cosmos_msgs(
        messages: Vec<CosmosMsg>,
        encoding: &TxEncoding,
        memo: Option<String>,
        ica_address: &str,
    ) -> StdResult<Self> {
        match encoding {
            TxEncoding::Protobuf => {
                let proto_anys = messages
                    .into_iter()
                    .map(|msg| -> StdResult<cosmos_sdk_proto::Any> {
                        convert_to_proto_any(msg, ica_address.to_string())
                            .map_err(|e| StdError::generic_err(e.to_string()))
                    })
                    .collect::<StdResult<Vec<cosmos_sdk_proto::Any>>>()?;

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

    use cosmos_sdk_proto::cosmos::base::abci::v1beta1::TxMsgData;
    use prost::Message;

    use crate::types::ContractError;

    #[cfg(feature = "query")]
    use crate::types::query_msg;

    use super::{cw_serde, StdError};

    /// `Data` is the response to an ibc packet. It either contains a result or an error.
    #[cw_serde]
    pub enum Data {
        /// Result is the result of a successful transaction.
        Result(Binary),
        /// Error is the error message of a failed transaction.
        /// It is a string of the error message (not base64 encoded).
        Error(String),
    }

    impl Data {
        /// `to_tx_msg_data` converts the acknowledgement to a [`TxMsgData`].
        ///
        /// # Errors
        /// Returns an error if the acknowledgement is an error or if the data cannot be decoded.
        pub fn to_tx_msg_data(&self) -> Result<TxMsgData, ContractError> {
            match self {
                Self::Result(data) => Ok(TxMsgData::decode(data.as_slice())?),
                Self::Error(err) => Err(StdError::generic_err(err))?,
            }
        }

        /// `decode_module_query_safe_resp` decodes the acknowledgement at the given index to a [`query_msg::proto::MsgModuleQuerySafeResponse`].
        ///
        /// # Errors
        /// Returns an error if the acknowledgement is an error or if the data at the index cannot be decoded.
        #[cfg(feature = "query")]
        pub fn decode_module_query_safe_resp(
            &self,
            index: usize,
        ) -> Result<query_msg::proto::MsgModuleQuerySafeResponse, ContractError> {
            let tx_msg_data = self.to_tx_msg_data()?;
            let msg_resp = tx_msg_data.msg_responses.get(index).ok_or_else(|| {
                StdError::generic_err("no MsgData found at the given index".to_string())
            })?;

            Ok(query_msg::proto::MsgModuleQuerySafeResponse::decode(
                msg_resp.value.as_slice(),
            )?)
        }

        /// `decode_module_query_safe_resp` decodes the acknowledgement at the last index to a [`query_msg::proto::MsgModuleQuerySafeResponse`].
        /// This is a convenience function since the contract only sends one query at the last index.
        ///
        /// # Errors
        /// Returns an error if the acknowledgement is an error or if the data at the index cannot be decoded.
        #[cfg(feature = "export")]
        pub fn decode_module_query_safe_resp_last_index(
            &self,
        ) -> Result<query_msg::proto::MsgModuleQuerySafeResponse, ContractError> {
            let tx_msg_data = self.to_tx_msg_data()?;
            let msg_resp = tx_msg_data.msg_responses.last().ok_or_else(|| {
                StdError::generic_err("no MsgData found at the given index".to_string())
            })?;

            Ok(query_msg::proto::MsgModuleQuerySafeResponse::decode(
                msg_resp.value.as_slice(),
            )?)
        }
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
