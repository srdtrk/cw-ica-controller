//! This module contains the helpers to convert [`CosmosMsg`] to [`cosmos_sdk_proto::Any`] or json string.

use cosmos_sdk_proto::{prost::EncodeError, Any};
use cosmwasm_std::{BankMsg, Coin, CosmosMsg, IbcMsg};

/// `convert_to_proto_any` converts a [`CosmosMsg`] to a [`cosmos_sdk_proto::Any`].
///
/// `from_address` is not used in [`CosmosMsg::Stargate`]
///
/// # Errors
///
/// Returns an error on serialization failure.
///
/// # Panics
///
/// Panics if the [`CosmosMsg`] is not supported.
///
/// ## List of supported [`CosmosMsg`]
///
/// - [`CosmosMsg::Stargate`]
/// - [`CosmosMsg::Bank`] with [`BankMsg::Send`]
/// - [`CosmosMsg::Ibc`] with [`IbcMsg::Transfer`]
/// - [`CosmosMsg::Wasm`] with [`cosmwasm_std::WasmMsg::Execute`]
/// - [`CosmosMsg::Wasm`] with [`cosmwasm_std::WasmMsg::Instantiate`]
/// - [`CosmosMsg::Wasm`] with [`cosmwasm_std::WasmMsg::Instantiate2`]
/// - [`CosmosMsg::Wasm`] with [`cosmwasm_std::WasmMsg::Migrate`]
/// - [`CosmosMsg::Wasm`] with [`cosmwasm_std::WasmMsg::UpdateAdmin`]
/// - [`CosmosMsg::Wasm`] with [`cosmwasm_std::WasmMsg::ClearAdmin`]
/// - [`CosmosMsg::Gov`] with [`cosmwasm_std::GovMsg::Vote`]
/// - [`CosmosMsg::Gov`] with [`cosmwasm_std::GovMsg::VoteWeighted`]
/// - [`CosmosMsg::Staking`] with [`cosmwasm_std::StakingMsg::Delegate`]
/// - [`CosmosMsg::Staking`] with [`cosmwasm_std::StakingMsg::Undelegate`]
/// - [`CosmosMsg::Staking`] with [`cosmwasm_std::StakingMsg::Redelegate`]
/// - [`CosmosMsg::Distribution`] with [`cosmwasm_std::DistributionMsg::WithdrawDelegatorReward`]
/// - [`CosmosMsg::Distribution`] with [`cosmwasm_std::DistributionMsg::SetWithdrawAddress`]
pub fn convert_to_proto_any(msg: CosmosMsg, from_address: String) -> Result<Any, EncodeError> {
    match msg {
        CosmosMsg::Stargate { type_url, value } => Ok(Any {
            type_url,
            value: value.to_vec(),
        }),
        CosmosMsg::Bank(bank_msg) => convert_to_any::bank(bank_msg, from_address),
        CosmosMsg::Ibc(ibc_msg) => convert_to_any::ibc(ibc_msg, from_address),
        CosmosMsg::Wasm(wasm_msg) => convert_to_any::wasm(wasm_msg, from_address),
        CosmosMsg::Gov(gov_msg) => Ok(convert_to_any::gov(gov_msg, from_address)),
        #[cfg(feature = "staking")]
        CosmosMsg::Staking(staking_msg) => convert_to_any::staking(staking_msg, from_address),
        #[cfg(feature = "staking")]
        CosmosMsg::Distribution(distribution_msg) => {
            convert_to_any::distribution(distribution_msg, from_address)
        }
        _ => panic!("Unsupported CosmosMsg"),
    }
}

mod convert_to_any {
    use cosmos_sdk_proto::{
        cosmos::{
            bank::v1beta1::MsgSend,
            base::v1beta1::Coin as ProtoCoin,
            gov::v1::{MsgVoteWeighted, WeightedVoteOption as ProtoWeightedVoteOption},
            gov::v1beta1::{MsgVote, VoteOption as ProtoVoteOption},
        },
        cosmwasm::wasm::v1::{
            MsgClearAdmin, MsgExecuteContract, MsgInstantiateContract, MsgInstantiateContract2,
            MsgMigrateContract, MsgUpdateAdmin,
        },
        ibc::{applications::transfer::v1::MsgTransfer, core::client::v1::Height},
        prost::EncodeError,
        traits::Message,
        Any,
    };

    use cosmwasm_std::{BankMsg, GovMsg, IbcMsg, VoteOption, WasmMsg};
    #[cfg(feature = "staking")]
    use cosmwasm_std::{DistributionMsg, StakingMsg};

    pub fn bank(msg: BankMsg, from_address: String) -> Result<Any, EncodeError> {
        match msg {
            BankMsg::Send { to_address, amount } => Any::from_msg(&MsgSend {
                from_address,
                to_address,
                amount: amount
                    .into_iter()
                    .map(|coin| ProtoCoin {
                        denom: coin.denom,
                        amount: coin.amount.to_string(),
                    })
                    .collect(),
            }),
            _ => panic!("Unsupported BankMsg"),
        }
    }

    pub fn ibc(msg: IbcMsg, sender: String) -> Result<Any, EncodeError> {
        match msg {
            IbcMsg::Transfer {
                channel_id,
                to_address,
                amount,
                timeout,
            } => Any::from_msg(&MsgTransfer {
                source_port: "transfer".to_string(),
                source_channel: channel_id,
                token: Some(ProtoCoin {
                    denom: amount.denom,
                    amount: amount.amount.to_string(),
                }),
                sender,
                receiver: to_address,
                timeout_height: timeout.block().map(|block| Height {
                    revision_number: block.revision,
                    revision_height: block.height,
                }),
                timeout_timestamp: timeout.timestamp().map_or(0, |timestamp| timestamp.nanos()),
            }),
            _ => panic!("Unsupported IbcMsg"),
        }
    }

    pub fn wasm(msg: WasmMsg, sender: String) -> Result<Any, EncodeError> {
        match msg {
            WasmMsg::Execute {
                contract_addr,
                msg,
                funds,
            } => Any::from_msg(&MsgExecuteContract {
                sender,
                contract: contract_addr,
                msg: msg.to_vec(),
                funds: funds
                    .into_iter()
                    .map(|coin| ProtoCoin {
                        denom: coin.denom,
                        amount: coin.amount.to_string(),
                    })
                    .collect(),
            }),
            WasmMsg::Instantiate {
                admin,
                code_id,
                msg,
                funds,
                label,
            } => Any::from_msg(&MsgInstantiateContract {
                admin: admin.unwrap_or_default(),
                sender,
                code_id,
                msg: msg.to_vec(),
                funds: funds
                    .into_iter()
                    .map(|coin| ProtoCoin {
                        denom: coin.denom,
                        amount: coin.amount.to_string(),
                    })
                    .collect(),
                label,
            }),
            WasmMsg::Migrate {
                contract_addr,
                new_code_id,
                msg,
            } => Any::from_msg(&MsgMigrateContract {
                sender,
                contract: contract_addr,
                code_id: new_code_id,
                msg: msg.to_vec(),
            }),
            WasmMsg::UpdateAdmin {
                contract_addr,
                admin,
            } => Any::from_msg(&MsgUpdateAdmin {
                sender,
                new_admin: admin,
                contract: contract_addr,
            }),
            WasmMsg::ClearAdmin { contract_addr } => Any::from_msg(&MsgClearAdmin {
                sender,
                contract: contract_addr,
            }),
            WasmMsg::Instantiate2 {
                admin,
                code_id,
                label,
                msg,
                funds,
                salt,
            } => {
                let proto_msg = MsgInstantiateContract2 {
                    sender,
                    admin: admin.unwrap_or_default(),
                    code_id,
                    label,
                    msg: msg.to_vec(),
                    funds: funds
                        .into_iter()
                        .map(|coin| ProtoCoin {
                            denom: coin.denom,
                            amount: coin.amount.to_string(),
                        })
                        .collect(),
                    salt: salt.to_vec(),
                    fix_msg: false,
                };

                // TODO: use Any::from_msg after cosmos-sdk-proto > 0.20.0
                Ok(Any {
                    type_url: "/cosmwasm.wasm.v1.MsgInstantiateContract2".to_string(),
                    value: proto_msg.encode_to_vec(),
                })
            }
            _ => panic!("Unsupported WasmMsg"),
        }
    }

    pub fn gov(msg: GovMsg, voter: String) -> Any {
        const fn convert_to_proto_vote_option(option: &VoteOption) -> ProtoVoteOption {
            match option {
                VoteOption::Yes => ProtoVoteOption::Yes,
                VoteOption::No => ProtoVoteOption::No,
                VoteOption::Abstain => ProtoVoteOption::Abstain,
                VoteOption::NoWithVeto => ProtoVoteOption::NoWithVeto,
            }
        }

        match msg {
            GovMsg::Vote { proposal_id, vote } => {
                let value = MsgVote {
                    voter,
                    proposal_id,
                    option: convert_to_proto_vote_option(&vote) as i32,
                };

                // TODO: use Any::from_msg when cosmos-sdk-proto is > 0.20.0
                Any {
                    type_url: "/cosmos.gov.v1beta1.MsgVote".to_string(),
                    value: value.encode_to_vec(),
                }
            }
            GovMsg::VoteWeighted {
                proposal_id,
                options,
            } => {
                let options: Vec<ProtoWeightedVoteOption> = options
                    .into_iter()
                    .map(|weighted_option| -> ProtoWeightedVoteOption {
                        ProtoWeightedVoteOption {
                            weight: weighted_option.weight.to_string(),
                            option: convert_to_proto_vote_option(&weighted_option.option) as i32,
                        }
                    })
                    .collect();

                let value = MsgVoteWeighted {
                    proposal_id,
                    voter,
                    options,
                    metadata: String::new(),
                };

                Any {
                    type_url: "/cosmos.gov.v1.MsgVoteWeighted".to_string(),
                    value: value.encode_to_vec(),
                }
            }
        }
    }

    #[cfg(feature = "staking")]
    pub fn staking(msg: StakingMsg, delegator_address: String) -> Result<Any, EncodeError> {
        use cosmos_sdk_proto::cosmos::staking::v1beta1::{
            MsgBeginRedelegate, MsgDelegate, MsgUndelegate,
        };

        match msg {
            StakingMsg::Delegate { validator, amount } => Any::from_msg(&MsgDelegate {
                delegator_address,
                validator_address: validator,
                amount: Some(ProtoCoin {
                    denom: amount.denom,
                    amount: amount.amount.to_string(),
                }),
            }),
            StakingMsg::Undelegate { validator, amount } => Any::from_msg(&MsgUndelegate {
                delegator_address,
                validator_address: validator,
                amount: Some(ProtoCoin {
                    denom: amount.denom,
                    amount: amount.amount.to_string(),
                }),
            }),
            StakingMsg::Redelegate {
                src_validator,
                dst_validator,
                amount,
            } => Any::from_msg(&MsgBeginRedelegate {
                delegator_address,
                validator_src_address: src_validator,
                validator_dst_address: dst_validator,
                amount: Some(ProtoCoin {
                    denom: amount.denom,
                    amount: amount.amount.to_string(),
                }),
            }),
            _ => panic!("Unsupported StakingMsg"),
        }
    }

    #[cfg(feature = "staking")]
    pub fn distribution(
        msg: DistributionMsg,
        delegator_address: String,
    ) -> Result<Any, EncodeError> {
        use cosmos_sdk_proto::cosmos::distribution::v1beta1::{
            MsgSetWithdrawAddress, MsgWithdrawDelegatorReward,
        };

        match msg {
            DistributionMsg::WithdrawDelegatorReward { validator } => {
                Any::from_msg(&MsgWithdrawDelegatorReward {
                    delegator_address,
                    validator_address: validator,
                })
            }
            DistributionMsg::SetWithdrawAddress { address } => {
                Any::from_msg(&MsgSetWithdrawAddress {
                    delegator_address,
                    withdraw_address: address,
                })
            }
            _ => panic!("Unsupported DistributionMsg"),
        }
    }
}

/// `convert_to_proto3json` converts a [`CosmosMsg`] to a json string formatted with
/// [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) encoding format.
///
/// # Panics
/// Panics if the [`CosmosMsg`] is not supported.
/// Notably, [`CosmosMsg::Stargate`] and [`CosmosMsg::Wasm`] are not supported.
///
/// ## List of supported [`CosmosMsg`]
///
/// - [`CosmosMsg::Bank`] with [`BankMsg::Send`]
/// - [`CosmosMsg::Ibc`] with [`IbcMsg::Transfer`]
/// - [`CosmosMsg::Gov`] with [`cosmwasm_std::GovMsg::Vote`]
/// - [`CosmosMsg::Gov`] with [`cosmwasm_std::GovMsg::VoteWeighted`]
/// - [`CosmosMsg::Staking`] with [`cosmwasm_std::StakingMsg::Delegate`]
/// - [`CosmosMsg::Staking`] with [`cosmwasm_std::StakingMsg::Undelegate`]
/// - [`CosmosMsg::Staking`] with [`cosmwasm_std::StakingMsg::Redelegate`]
/// - [`CosmosMsg::Distribution`] with [`cosmwasm_std::DistributionMsg::WithdrawDelegatorReward`]
/// - [`CosmosMsg::Distribution`] with [`cosmwasm_std::DistributionMsg::SetWithdrawAddress`]
#[must_use]
pub fn convert_to_proto3json(msg: CosmosMsg, from_address: String) -> String {
    match msg {
        CosmosMsg::Bank(msg) => convert_to_json::bank(msg, from_address),
        CosmosMsg::Ibc(msg) => convert_to_json::ibc(msg, from_address),
        CosmosMsg::Gov(msg) => convert_to_json::gov(msg, from_address),
        #[cfg(feature = "staking")]
        CosmosMsg::Staking(msg) => convert_to_json::staking(msg, from_address),
        #[cfg(feature = "staking")]
        CosmosMsg::Distribution(msg) => convert_to_json::distribution(msg, from_address),
        _ => panic!("Unsupported CosmosMsg"),
    }
}

mod convert_to_json {
    #[cfg(feature = "staking")]
    use cosmwasm_std::{DistributionMsg, StakingMsg};
    use cosmwasm_std::{GovMsg, VoteOption};

    use super::{BankMsg, Coin, IbcMsg};

    pub fn bank(msg: BankMsg, from_address: String) -> String {
        match msg {
            BankMsg::Send { to_address, amount } => CosmosMsgProto3JsonSerializer::Send {
                from_address,
                to_address,
                amount,
            },
            _ => panic!("Unsupported BankMsg"),
        }
        .to_string()
    }

    pub fn ibc(msg: IbcMsg, sender: String) -> String {
        match msg {
            IbcMsg::Transfer {
                channel_id,
                to_address,
                amount,
                timeout,
            } => CosmosMsgProto3JsonSerializer::Transfer {
                source_port: "transfer".to_string(),
                source_channel: channel_id,
                token: amount,
                sender,
                receiver: to_address,
                timeout_height: timeout.block().map_or(
                    Height {
                        revision_number: 0,
                        revision_height: 0,
                    },
                    |block| Height {
                        revision_number: block.revision,
                        revision_height: block.height,
                    },
                ),
                timeout_timestamp: timeout.timestamp().map_or(0, |timestamp| timestamp.nanos()),
                memo: None,
            },
            _ => panic!("Unsupported IbcMsg"),
        }
        .to_string()
    }

    pub fn gov(msg: GovMsg, voter: String) -> String {
        const fn convert_to_u64(option: &VoteOption) -> u64 {
            match option {
                VoteOption::Yes => 1,
                VoteOption::Abstain => 2,
                VoteOption::No => 3,
                VoteOption::NoWithVeto => 4,
            }
        }

        match msg {
            GovMsg::Vote { proposal_id, vote } => CosmosMsgProto3JsonSerializer::Vote {
                voter,
                proposal_id,
                option: convert_to_u64(&vote),
            },
            GovMsg::VoteWeighted {
                proposal_id,
                options,
            } => CosmosMsgProto3JsonSerializer::VoteWeighted {
                proposal_id,
                voter,
                options: options
                    .into_iter()
                    .map(|weighted_option| -> WeightedVoteOption {
                        WeightedVoteOption {
                            weight: weighted_option.weight.to_string(),
                            option: convert_to_u64(&weighted_option.option),
                        }
                    })
                    .collect(),
            },
        }
        .to_string()
    }

    #[cfg(feature = "staking")]
    pub fn staking(msg: StakingMsg, delegator_address: String) -> String {
        match msg {
            StakingMsg::Delegate { validator, amount } => CosmosMsgProto3JsonSerializer::Delegate {
                delegator_address,
                validator_address: validator,
                amount,
            },
            StakingMsg::Undelegate { validator, amount } => {
                CosmosMsgProto3JsonSerializer::Undelegate {
                    delegator_address,
                    validator_address: validator,
                    amount,
                }
            }
            StakingMsg::Redelegate {
                src_validator,
                dst_validator,
                amount,
            } => CosmosMsgProto3JsonSerializer::Redelegate {
                delegator_address,
                validator_src_address: src_validator,
                validator_dst_address: dst_validator,
                amount,
            },
            _ => panic!("Unsupported StakingMsg"),
        }
        .to_string()
    }

    #[cfg(feature = "staking")]
    pub fn distribution(msg: DistributionMsg, delegator_address: String) -> String {
        match msg {
            DistributionMsg::WithdrawDelegatorReward { validator } => {
                CosmosMsgProto3JsonSerializer::WithdrawDelegatorReward {
                    delegator_address,
                    validator_address: validator,
                }
            }
            DistributionMsg::SetWithdrawAddress { address } => {
                CosmosMsgProto3JsonSerializer::SetWithdrawAddress {
                    delegator_address,
                    withdraw_address: address,
                }
            }
            _ => panic!("Unsupported DistributionMsg"),
        }
        .to_string()
    }

    /// `CosmosMsgProto3JsonSerializer` is a list of Cosmos messages that can be sent to the ICA host if the channel handshake is
    /// completed with the [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) encoding format.
    ///
    /// This enum corresponds to the [Any](https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/codec/types/any.go#L11-L52)
    /// type defined in the Cosmos SDK. The Any type is used to encode and decode Cosmos messages. It also has a built-in
    /// json codec. This enum is used to encode Cosmos messages using json so that they can be deserialized as an Any by
    /// the host chain using the Cosmos SDK's json codec.
    ///
    /// This enum does not derive [Deserialize](serde::Deserialize), see issue
    /// [#1443](https://github.com/CosmWasm/cosmwasm/issues/1443)
    #[derive(serde::Serialize, Clone, Debug, PartialEq, Eq)]
    #[cfg_attr(test, derive(serde::Deserialize))]
    #[serde(tag = "@type")]
    pub enum CosmosMsgProto3JsonSerializer {
        /// This is a Cosmos message to send tokens from one account to another.
        #[serde(rename = "/cosmos.bank.v1beta1.MsgSend")]
        Send {
            /// Sender's address.
            from_address: String,
            /// Recipient's address.
            to_address: String,
            /// Amount to send
            amount: Vec<Coin>,
        },
        /// This is an IBC transfer message.
        #[serde(rename = "/ibc.applications.transfer.v1.MsgTransfer")]
        Transfer {
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
            timeout_height: Height,
            /// Timeout timestamp. Disabled when set to 0.
            timeout_timestamp: u64,
            /// Optional memo.
            #[serde(skip_serializing_if = "Option::is_none")]
            memo: Option<String>,
        },
        /// This is a Cosmos message to vote on a governance proposal.
        #[serde(rename = "/cosmos.gov.v1beta1.MsgVote")]
        Vote {
            voter: String,
            proposal_id: u64,
            option: u64,
        },
        /// This is a Cosmos message to vote on a governance proposal with weighted votes.
        #[serde(rename = "/cosmos.gov.v1beta1.MsgVoteWeighted")]
        VoteWeighted {
            proposal_id: u64,
            voter: String,
            options: Vec<WeightedVoteOption>,
        },
        /// This is a Cosmos message to delegate tokens to a validator.
        #[cfg(feature = "staking")]
        #[serde(rename = "/cosmos.staking.v1beta1.MsgDelegate")]
        Delegate {
            delegator_address: String,
            validator_address: String,
            amount: Coin,
        },
        /// This is a Cosmos message to undelegate tokens from a validator.
        #[cfg(feature = "staking")]
        #[serde(rename = "/cosmos.staking.v1beta1.MsgUndelegate")]
        Undelegate {
            delegator_address: String,
            validator_address: String,
            amount: Coin,
        },
        /// This is a Cosmos message to redelegate tokens from one validator to another.
        #[cfg(feature = "staking")]
        #[serde(rename = "/cosmos.staking.v1beta1.MsgBeginRedelegate")]
        Redelegate {
            delegator_address: String,
            validator_src_address: String,
            validator_dst_address: String,
            amount: Coin,
        },
        /// This is a Cosmos message to withdraw rewards from a validator.
        #[cfg(feature = "staking")]
        #[serde(rename = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward")]
        WithdrawDelegatorReward {
            delegator_address: String,
            validator_address: String,
        },
        /// This is a Cosmos message to set the withdraw address for a delegator.
        #[cfg(feature = "staking")]
        #[serde(rename = "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress")]
        SetWithdrawAddress {
            delegator_address: String,
            withdraw_address: String,
        },
    }

    /// This is a helper struct to serialize the `WeightedVoteOption` struct in [`CosmosMsgProto3JsonSerializer`].
    #[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq, Eq)]
    pub struct WeightedVoteOption {
        pub option: u64,
        pub weight: String,
    }

    /// This is a helper struct to serialize the `Height` struct in [`CosmosMsgProto3JsonSerializer`].
    #[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq, Eq)]
    pub struct Height {
        pub revision_number: u64,
        pub revision_height: u64,
    }

    impl ToString for CosmosMsgProto3JsonSerializer {
        fn to_string(&self) -> String {
            serde_json_wasm::to_string(self).unwrap()
        }
    }
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use cosmwasm_std::{coins, from_json, Decimal, Uint128, VoteOption, WeightedVoteOption};

    use crate::ibc::types::packet::IcaPacketData;

    use super::convert_to_json::CosmosMsgProto3JsonSerializer;

    #[test]
    fn test_json_support() {
        #[derive(serde::Serialize, serde::Deserialize)]
        struct TestCosmosTx {
            pub messages: Vec<CosmosMsgProto3JsonSerializer>,
        }

        let packet_from_string = IcaPacketData::from_json_strings(
            &[r#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#.to_string()], None);

        let packet_data = packet_from_string.data;
        let cosmos_tx: TestCosmosTx = from_json(packet_data).unwrap();

        let expected = CosmosMsgProto3JsonSerializer::Send {
            from_address: "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk".to_string(),
            to_address: "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk".to_string(),
            amount: coins(5000, "stake".to_string()),
        };

        assert_eq!(expected, cosmos_tx.messages[0]);
    }

    #[test]
    fn test_weighted_vote_option() {
        let test_msg = r#"{"option":"yes","weight":"0.5"}"#;

        let vote_option = serde_json_wasm::from_str::<WeightedVoteOption>(test_msg).unwrap();

        assert_eq!(
            vote_option,
            WeightedVoteOption {
                option: VoteOption::Yes,
                weight: Decimal::from_ratio(
                    Uint128::from_str("1").unwrap(),
                    Uint128::from_str("2").unwrap()
                ),
            }
        );
        assert_eq!("0.5".to_string(), vote_option.weight.to_string());
    }
}
