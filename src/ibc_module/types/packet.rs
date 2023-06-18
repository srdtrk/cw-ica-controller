use cosmwasm_schema::cw_serde;

/// InterchainAccountPacketData is comprised of a raw transaction, type of transaction and optional memo field.
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
}

pub mod ica_cosmos_tx {
    use crate::types::ContractError;

    use super::*;

    /// IcaCosmosTx is a list of Cosmos messages that is sent to the ICA host.
    ///
    /// ## Example
    ///
    /// The messages must be serialized as JSON strings in the format expected by the ICA host.
    /// The following is an example of a serialized IcaCosmosTx with one legacy gov proposal message:
    ///
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
    ///
    /// We leave it to the user to serialize the messages in the format expected by the ICA host.
    /// For example, this can be achieved via custom user sent messages such as
    /// [ExecuteMsg::SendCustomIcaMessages](crate::types::msg::ExecuteMsg::SendCustomIcaMessages)
    /// or by using predefined messages such as [CosmosMessages](crate::types::cosmos_msg::CosmosMessages).
    #[cw_serde]
    pub struct IcaCosmosTx {
        pub messages: Vec<serde_json::Value>,
    }

    impl IcaCosmosTx {
        /// Creates a new IcaCosmosTx
        pub fn new(messages: Vec<serde_json::Value>) -> Self {
            Self { messages }
        }

        /// Creates a new IcaCosmosTx from a list of JSON strings
        pub fn from_strings(messages: Vec<impl Into<String>>) -> Result<Self, ContractError> {
            let maybe_json_messages: Result<Vec<serde_json::Value>, _> = messages
                .into_iter()
                .map(|msg| serde_json_wasm::from_str(&msg.into()))
                .collect();
            Ok(Self::new(maybe_json_messages?))
        }
    }
}
