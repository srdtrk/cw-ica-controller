//! This module defines the [`ExampleCosmosMessages`] enum which is used to encode Cosmos messages as Any's in
//! [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) encoding format.

use cosmwasm_std::Coin;

/// ExampleCosmosMessages is a list of Cosmos messages that can be sent to the ICA host if the channel handshake is
/// completed with the [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) encoding format.
///
/// This enum corresponds to the [Any](https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/codec/types/any.go#L11-L52)
/// type defined in the Cosmos SDK. The Any type is used to encode and decode Cosmos messages. It also has a built-in
/// json codec. This enum is used to encode Cosmos messages using json so that they can be deserialized as an Any by
/// the host chain using the Cosmos SDK's json codec.
///
/// In general, this ICA controller should be used with custom messages and **not with the messages defined here**.
/// The messages defined here are to demonstrate how an ICA controller can be used with registered
/// ExampleCosmosMessages (in case the contract is a DAO with **predefined actions**)
///
/// This enum does not derive [Deserialize](serde::Deserialize), see issue
/// [#1443](https://github.com/CosmWasm/cosmwasm/issues/1443)
#[derive(serde::Serialize, Clone, Debug, PartialEq)]
#[cfg_attr(test, derive(serde::Deserialize))]
#[serde(tag = "@type")]
pub enum ExampleCosmosMessages {
    /// This is a Cosmos message to send tokens from one account to another.
    #[serde(rename = "/cosmos.bank.v1beta1.MsgSend")]
    MsgSend {
        /// Sender's address.
        from_address: String,
        /// Recipient's address.
        to_address: String,
        /// Amount to send
        amount: Vec<Coin>,
    },
    /// This is a Cosmos message to delegate tokens to a validator.
    #[serde(rename = "/cosmos.staking.v1beta1.MsgDelegate")]
    MsgDelegate {
        /// Delegator's address.
        delegator_address: String,
        /// Validator's address.
        validator_address: String,
        /// Amount to delegate.
        amount: Coin,
    },
    /// This is a Cosmos message to vote on a proposal.
    #[serde(rename = "/cosmos.gov.v1beta1.MsgVote")]
    MsgVote {
        /// Voter's address.
        voter: String,
        /// Proposal's ID.
        proposal_id: u64,
        /// Vote option.
        option: u32,
    },
    /// This is a legacy submit governance proposal message.
    #[serde(rename = "/cosmos.gov.v1beta1.MsgSubmitProposal")]
    MsgSubmitProposalLegacy {
        /// Content is another Cosmos message.
        content: Box<ExampleCosmosMessages>,
        /// Initial deposit to the proposal.
        initial_deposit: Vec<Coin>,
        /// Proposer's address. (In this case, ICA address)
        proposer: String,
    },
    /// This is a text governance proposal message.
    #[serde(rename = "/cosmos.gov.v1beta1.TextProposal")]
    TextProposal {
        /// Proposal's title
        title: String,
        /// Proposal's description
        description: String,
    },
    /// This is a Cosmos message to deposit tokens to a proposal.
    #[serde(rename = "/cosmos.gov.v1beta1.MsgDeposit")]
    MsgDeposit {
        /// Proposal's ID.
        proposal_id: u64,
        /// Depositor's address. (In this case, ICA address)
        depositor: String,
        /// Amount to deposit.
        amount: Vec<Coin>,
    },
    /// This is an IBC transfer message.
    #[serde(rename = "/ibc.applications.transfer.v1.MsgTransfer")]
    MsgTransfer {
        /// Source port.
        source_port: String,
        /// Source channel id.
        source_channel: String,
        /// Amount to transfer.
        token: Coin,
        /// Sender's address. (In this case, ICA address)
        sender: String,
        /// Recipient's address.
        receiver: String,
        /// Timeout height. Disabled when set to 0.
        timeout_height: msg_transfer::Height,
        /// Timeout timestamp. Disabled when set to 0.
        timeout_timestamp: u64,
        /// Optional memo.
        #[serde(skip_serializing_if = "Option::is_none")]
        memo: Option<String>,
    },
}

impl ToString for ExampleCosmosMessages {
    fn to_string(&self) -> String {
        serde_json_wasm::to_string(self).unwrap()
    }
}

mod msg_transfer {
    #[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
    pub struct Height {
        pub revision_number: u64,
        pub revision_height: u64,
    }
}
