use cosmwasm_std::Coin;
use serde::Serialize;

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

pub mod msg_transfer {
    use super::*;

    #[derive(Serialize, Clone, Debug, PartialEq)]
    pub struct Height {
        pub revision_number: u64,
        pub revision_height: u64,
    }
}
