use cosmwasm_schema::cw_serde;
use cosmwasm_std::{to_binary, Env, IbcMsg, IbcTimeout};

use crate::types::ContractError;

/// DEFAULT_TIMEOUT_SECONDS is the default timeout for [`InterchainAccountPacketData`]
pub const DEFAULT_TIMEOUT_SECONDS: u64 = 600;

/// InterchainAccountPacketData is comprised of a raw transaction, type of transaction and optional memo field.
/// Currently, the host only supports serialized [`IcaCosmosTx`](ica_cosmos_tx::IcaCosmosTx) messages as raw transactions.
#[cw_serde]
pub struct InterchainAccountPacketData {
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
    /// Currently, the host only supports serialized [`IcaCosmosTx`](ica_cosmos_tx::IcaCosmosTx) messages.
    pub data: Vec<u8>,
    /// Memo is an optional field that can be used to attach a memo to a transaction.
    /// It is also caught by some ibc middleware to perform additional actions.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub memo: Option<String>,
}

impl InterchainAccountPacketData {
    /// Creates a new InterchainAccountPacketData
    pub fn new(data: Vec<u8>, memo: Option<String>) -> Self {
        Self {
            r#type: 1,
            data,
            memo,
        }
    }

    /// Creates a new InterchainAccountPacketData from a list of JSON strings
    ///
    /// The messages must be serialized as JSON strings in the format expected by the ICA host.
    /// The following is an example of a serialized [`IcaCosmosTx`](ica_cosmos_tx::IcaCosmosTx) with one legacy gov proposal message:
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
    pub fn from_strings(
        messages: Vec<String>,
        memo: Option<String>,
    ) -> Result<Self, ContractError> {
        let combined_messages = messages.join(", ");
        let json_txs = format!(r#"{{"messages": [{}]}}"#, combined_messages);
        let data = json_txs.into_bytes();
        Ok(Self::new(data, memo))
    }

    /// Creates an IBC SendPacket message from the InterchainAccountPacketData
    pub fn to_ibc_msg(
        &self,
        env: &Env,
        channel_id: impl Into<String>,
        timeout_seconds: Option<u64>,
    ) -> Result<IbcMsg, ContractError> {
        let timeout_timestamp = env
            .block
            .time
            .plus_seconds(timeout_seconds.unwrap_or(DEFAULT_TIMEOUT_SECONDS));
        Ok(IbcMsg::SendPacket {
            channel_id: channel_id.into(),
            data: to_binary(&self)?,
            timeout: IbcTimeout::with_timestamp(timeout_timestamp),
        })
    }
}

pub mod acknowledgement {
    use cosmwasm_std::Binary;

    use super::*;

    /// Acknowledgement is the response to an ibc packet. It either contains a result or an error.
    #[cw_serde]
    pub enum AcknowledgementData {
        /// Result is the result of a successful transaction.
        Result(Binary),
        /// Error is the error message of a failed transaction.
        /// It is a string of the error message (not base64 encoded).
        Error(String),
    }
}

#[cfg(test)]
mod tests {
    use acknowledgement::AcknowledgementData;
    use cosmwasm_std::{coins, from_binary, Binary};
    use serde::{Deserialize, Serialize};

    use crate::types::cosmos_msg::CosmosMessages;

    use super::*;

    #[test]
    fn test_packet_data() {
        #[derive(Serialize, Deserialize)]
        struct TestCosmosTx {
            pub messages: Vec<CosmosMessages>,
        }

        let packet_from_string = InterchainAccountPacketData::from_strings(
            vec![r#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#.to_string()], None).unwrap();

        let packet_data = packet_from_string.data;
        let cosmos_tx: TestCosmosTx = from_binary(&Binary(packet_data)).unwrap();

        let expected = CosmosMessages::MsgSend {
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
        let ack: AcknowledgementData = from_binary(&cw_success_binary).unwrap();
        assert_eq!(
            ack,
            AcknowledgementData::Result(Binary::from_base64("c3VjY2Vzcw==").unwrap())
        );

        // Test error:
        let error_bytes =
            br#"{"error":"ABCI code: 1: error handling packet: see events for details"}"#;
        let cw_error_binary = Binary(error_bytes.to_vec());
        let ack: AcknowledgementData = from_binary(&cw_error_binary).unwrap();
        assert_eq!(
            ack,
            AcknowledgementData::Error(
                "ABCI code: 1: error handling packet: see events for details".to_string()
            )
        );
    }
}
