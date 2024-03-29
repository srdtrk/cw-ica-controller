//! This module contains the helpers to convert [`CosmosMsg`] to [`cosmos_sdk_proto::Any`]
//! or a [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) string.

use cosmos_sdk_proto::{prost::EncodeError, Any};
use cosmwasm_std::CosmosMsg;

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
/// - [`CosmosMsg::Distribution`] with [`cosmwasm_std::DistributionMsg::FundCommunityPool`]
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
        match msg {
            StakingMsg::Delegate { validator, amount } => {
                Any::from_msg(&cosmos_sdk_proto::cosmos::staking::v1beta1::MsgDelegate {
                    delegator_address,
                    validator_address: validator,
                    amount: Some(ProtoCoin {
                        denom: amount.denom,
                        amount: amount.amount.to_string(),
                    }),
                })
            }
            StakingMsg::Undelegate { validator, amount } => {
                Any::from_msg(&cosmos_sdk_proto::cosmos::staking::v1beta1::MsgUndelegate {
                    delegator_address,
                    validator_address: validator,
                    amount: Some(ProtoCoin {
                        denom: amount.denom,
                        amount: amount.amount.to_string(),
                    }),
                })
            }
            StakingMsg::Redelegate {
                src_validator,
                dst_validator,
                amount,
            } => Any::from_msg(
                &cosmos_sdk_proto::cosmos::staking::v1beta1::MsgBeginRedelegate {
                    delegator_address,
                    validator_src_address: src_validator,
                    validator_dst_address: dst_validator,
                    amount: Some(ProtoCoin {
                        denom: amount.denom,
                        amount: amount.amount.to_string(),
                    }),
                },
            ),
            _ => panic!("Unsupported StakingMsg"),
        }
    }

    #[cfg(feature = "staking")]
    pub fn distribution(msg: DistributionMsg, sender: String) -> Result<Any, EncodeError> {
        match msg {
            DistributionMsg::WithdrawDelegatorReward { validator } => Any::from_msg(
                &cosmos_sdk_proto::cosmos::distribution::v1beta1::MsgWithdrawDelegatorReward {
                    delegator_address: sender,
                    validator_address: validator,
                },
            ),
            DistributionMsg::SetWithdrawAddress { address } => Any::from_msg(
                &cosmos_sdk_proto::cosmos::distribution::v1beta1::MsgSetWithdrawAddress {
                    delegator_address: sender,
                    withdraw_address: address,
                },
            ),
            DistributionMsg::FundCommunityPool { amount } => Any::from_msg(
                &cosmos_sdk_proto::cosmos::distribution::v1beta1::MsgFundCommunityPool {
                    depositor: sender,
                    amount: amount
                        .into_iter()
                        .map(|coin| ProtoCoin {
                            denom: coin.denom,
                            amount: coin.amount.to_string(),
                        })
                        .collect(),
                },
            ),
            _ => panic!("Unsupported DistributionMsg"),
        }
    }
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use cosmwasm_std::{Decimal, Uint128, VoteOption, WeightedVoteOption};

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
