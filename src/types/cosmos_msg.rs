use cosmwasm_std::Coin;
use serde::Serialize;

/// CosmosMessages is a list of Cosmos messages that can be sent to the ICA host.
///
/// In general, this ICA controller should be used with custom messages and **not with the
/// messages defined here**. The messages defined here are to demonstrate how an ICA controller
/// can be used with registered CosmosMessages (in case the contract is a DAO with **predefined actions**)
///
/// This enum does not derive Deserialize, see issue [#1443](https://github.com/CosmWasm/cosmwasm/issues/1443)
#[derive(Serialize, Clone, Debug, PartialEq)]
#[serde(tag = "@type")]
pub enum CosmosMessages {
    #[serde(rename = "/cosmos.bank.v1beta1.MsgSend")]
    MsgSend {
        /// Sender's address.
        from_address: String,
        /// Recipient's address.
        to_address: String,
        /// Amount to send
        amount: Vec<Coin>,
    },
    #[serde(rename = "/cosmos.staking.v1beta1.MsgDelegate")]
    MsgDelegate {
        /// Delegator's address.
        delegator_address: String,
        /// Validator's address.
        validator_address: String,
        /// Amount to delegate.
        amount: Coin,
    },
    #[serde(rename = "/cosmos.gov.v1beta1.MsgVote")]
    MsgVote {
        /// Voter's address.
        voter: String,
        /// Proposal's ID.
        proposal_id: u64,
        /// Vote option.
        option: u32,
    },
    /// This legacy submit proposal message is used to demonstrate how CosmosMessages
    /// can embed other CosmosMessages.
    #[serde(rename = "/cosmos.gov.v1beta1.MsgSubmitProposal")]
    MsgSubmitProposalLegacy {
        content: Box<CosmosMessages>,
        initial_deposit: Vec<Coin>,
        proposer: String,
    },
    #[serde(rename = "/cosmos.gov.v1beta1.TextProposal")]
    TextProposal { title: String, description: String },
    #[serde(rename = "/cosmos.gov.v1beta1.MsgDeposit")]
    MsgDeposit {
        proposal_id: u64,
        depositor: String,
        amount: Vec<Coin>,
    },
    #[serde(rename = "/ibc.applications.transfer.v1.MsgTransfer")]
    MsgTransfer {
        source_port: String,
        source_channel: String,
        token: Coin,
        sender: String,
        receiver: String,
        timeout_height: msg_transfer::Height,
        timeout_timestamp: u64,
    },
}

impl ToString for CosmosMessages {
    fn to_string(&self) -> String {
        serde_json_wasm::to_string(self).unwrap()
    }
}

pub mod msg_transfer {
    use super::*;

    #[derive(Serialize, Clone, Debug, PartialEq)]
    pub struct Height {
        pub revision_number: u64,
        pub revision_height: u64,
    }
}
