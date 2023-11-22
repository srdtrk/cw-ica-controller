//! # Packet
//!
//! This module contains the packet data to be send to the ica host and acknowledgement data types.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{to_json_binary, Env, IbcMsg, IbcTimeout, StdResult};

pub use cosmos_sdk_proto::ibc::applications::interchain_accounts::v1::CosmosTx;
use cosmos_sdk_proto::traits::Message;

/// `DEFAULT_TIMEOUT_SECONDS` is the default timeout for [`IcaPacketData`]
pub const DEFAULT_TIMEOUT_SECONDS: u64 = 600;

/// `IcaPacketData` is comprised of a raw transaction, type of transaction and optional memo field.
/// Currently, the host only supports [protobuf](super::metadata::TxEncoding::Protobuf) or
/// [proto3json](super::metadata::TxEncoding::Proto3Json) serialized Cosmos transactions.
///
/// If protobuf is used, then the raw transaction must encoded using
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

    /// Creates a new [`IcaPacketData`] from a list of JSON strings
    ///
    /// The messages must be serialized as JSON strings in the format expected by the ICA host.
    /// The following is an example with one legacy gov proposal message:
    ///
    /// ## Format
    ///
    /// ```json
    /// {
    ///   "messages": [
    ///     {
    ///       "@type": "/cosmos.gov.v1beta1.MsgSubmitProposal",
    ///       "content": {
    ///         "@type": "/cosmos.gov.v1beta1.TextProposal",
    ///         "title": "IBC Gov Proposal",
    ///         "description": "tokens for all!"
    ///       },
    ///       "initial_deposit": [{ "denom": "stake", "amount": "5000" }],
    ///       "proposer": "cosmos1k4epd6js8aa7fk4e5l7u6dwttxfarwu6yald9hlyckngv59syuyqnlqvk8"
    ///     }
    ///   ]
    /// }
    /// ```
    ///
    /// In this example, the proposer must be the ICA controller's address.
    #[must_use]
    pub fn from_json_strings(messages: &[String], memo: Option<String>) -> Self {
        let combined_messages = messages.join(", ");
        let json_txs = format!(r#"{{"messages": [{combined_messages}]}}"#);
        let data = json_txs.into_bytes();
        Self::new(data, memo)
    }

    /// Creates a new [`IcaPacketData`] from a list of [`cosmos_sdk_proto::Any`] messages
    #[must_use]
    pub fn from_proto_anys(messages: Vec<cosmos_sdk_proto::Any>, memo: Option<String>) -> Self {
        let cosmos_tx = CosmosTx { messages };
        let data = cosmos_tx.encode_to_vec();
        Self::new(data, memo)
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
    use cosmwasm_std::{coins, from_json, Binary};

    use crate::types::cosmos_msg::ExampleCosmosMessages;

    use super::*;

    #[test]
    fn test_packet_data() {
        #[derive(serde::Serialize, serde::Deserialize)]
        struct TestCosmosTx {
            pub messages: Vec<ExampleCosmosMessages>,
        }

        let packet_from_string = IcaPacketData::from_json_strings(
            &[r#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#.to_string()], None);

        let packet_data = packet_from_string.data;
        let cosmos_tx: TestCosmosTx = from_json(Binary(packet_data)).unwrap();

        let expected = ExampleCosmosMessages::MsgSend {
            from_address: "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk".to_string(),
            to_address: "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk".to_string(),
            amount: coins(5000, "stake".to_string()),
        };

        assert_eq!(expected, cosmos_tx.messages[0]);
    }

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
