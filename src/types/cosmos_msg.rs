//! This module contains the helpers to convert [`CosmosMsg`] to [`cosmos_sdk_proto::Any`] or json string.

use cosmos_sdk_proto::{
    cosmos::{bank::v1beta1::MsgSend, base::v1beta1::Coin as ProtoCoin},
    cosmwasm::wasm::v1::{
        MsgClearAdmin, MsgExecuteContract, MsgInstantiateContract, MsgMigrateContract,
        MsgUpdateAdmin,
    },
    ibc::{applications::transfer::v1::MsgTransfer, core::client::v1::Height},
    prost::EncodeError,
    Any,
};
use cosmwasm_std::{BankMsg, Coin, CosmosMsg, IbcMsg, WasmMsg};

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
/// - [`CosmosMsg::Wasm`] with [`WasmMsg::Execute`]
/// - [`CosmosMsg::Wasm`] with [`WasmMsg::Instantiate`]
/// - [`CosmosMsg::Wasm`] with [`WasmMsg::Migrate`]
/// - [`CosmosMsg::Wasm`] with [`WasmMsg::UpdateAdmin`]
/// - [`CosmosMsg::Wasm`] with [`WasmMsg::ClearAdmin`]
pub fn convert_to_proto_any(msg: CosmosMsg, from_address: String) -> Result<Any, EncodeError> {
    match msg {
        CosmosMsg::Stargate { type_url, value } => Ok(Any {
            type_url,
            value: value.to_vec(),
        }),
        CosmosMsg::Bank(BankMsg::Send { to_address, amount }) => Any::from_msg(&MsgSend {
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
        CosmosMsg::Ibc(IbcMsg::Transfer {
            channel_id,
            to_address,
            amount,
            timeout,
        }) => Any::from_msg(&MsgTransfer {
            source_port: "transfer".to_string(),
            source_channel: channel_id,
            token: Some(ProtoCoin {
                denom: amount.denom,
                amount: amount.amount.to_string(),
            }),
            sender: from_address,
            receiver: to_address,
            timeout_height: timeout.block().map(|block| Height {
                revision_number: block.revision,
                revision_height: block.height,
            }),
            timeout_timestamp: timeout.timestamp().map_or(0, |timestamp| timestamp.nanos()),
        }),
        CosmosMsg::Wasm(WasmMsg::Execute {
            contract_addr,
            msg,
            funds,
        }) => Any::from_msg(&MsgExecuteContract {
            sender: from_address,
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
        CosmosMsg::Wasm(WasmMsg::Instantiate {
            admin,
            code_id,
            msg,
            funds,
            label,
        }) => Any::from_msg(&MsgInstantiateContract {
            admin: admin.unwrap_or_default(),
            sender: from_address,
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
        CosmosMsg::Wasm(WasmMsg::Migrate {
            contract_addr,
            new_code_id,
            msg,
        }) => Any::from_msg(&MsgMigrateContract {
            sender: from_address,
            contract: contract_addr,
            code_id: new_code_id,
            msg: msg.to_vec(),
        }),
        CosmosMsg::Wasm(WasmMsg::UpdateAdmin {
            contract_addr,
            admin,
        }) => Any::from_msg(&MsgUpdateAdmin {
            sender: from_address,
            new_admin: admin,
            contract: contract_addr,
        }),
        CosmosMsg::Wasm(WasmMsg::ClearAdmin { contract_addr }) => Any::from_msg(&MsgClearAdmin {
            sender: from_address,
            contract: contract_addr,
        }),
        _ => panic!("Unsupported CosmosMsg"),
    }
}

/// `convert_to_proto3json` converts a [`CosmosMsg`] to a json string formatted with
/// [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) encoding format.
///
/// # Panics
/// Panics if the [`CosmosMsg`] is not supported.
///
/// ## List of supported [`CosmosMsg`]
///
/// - [`CosmosMsg::Bank`] with [`BankMsg::Send`]
/// - [`CosmosMsg::Ibc`] with [`IbcMsg::Transfer`]
#[must_use]
pub fn convert_to_proto3json(msg: CosmosMsg, from_address: String) -> String {
    match msg {
        CosmosMsg::Bank(BankMsg::Send { to_address, amount }) => {
            JsonSupportedCosmosMessages::MsgSend {
                from_address,
                to_address,
                amount,
            }
            .to_string()
        }
        CosmosMsg::Ibc(IbcMsg::Transfer {
            channel_id,
            to_address,
            amount,
            timeout,
        }) => JsonSupportedCosmosMessages::MsgTransfer {
            source_port: "transfer".to_string(),
            source_channel: channel_id,
            token: amount,
            sender: from_address,
            receiver: to_address,
            timeout_height: timeout.block().map_or(
                msg_transfer::Height {
                    revision_number: 0,
                    revision_height: 0,
                },
                |block| msg_transfer::Height {
                    revision_number: block.revision,
                    revision_height: block.height,
                },
            ),
            timeout_timestamp: timeout.timestamp().map_or(0, |timestamp| timestamp.nanos()),
            memo: None,
        }
        .to_string(),
        _ => panic!("Unsupported CosmosMsg"),
    }
}

/// `JsonSupportedCosmosMessages` is a list of Cosmos messages that can be sent to the ICA host if the channel handshake is
/// completed with the [`proto3json`](crate::ibc::types::metadata::TxEncoding::Proto3Json) encoding format.
///
/// This enum corresponds to the [Any](https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/codec/types/any.go#L11-L52)
/// type defined in the Cosmos SDK. The Any type is used to encode and decode Cosmos messages. It also has a built-in
/// json codec. This enum is used to encode Cosmos messages using json so that they can be deserialized as an Any by
/// the host chain using the Cosmos SDK's json codec.
///
/// In general, this ICA controller should be used with custom messages and **not with the messages defined here**.
/// The messages defined here are to demonstrate how an ICA controller can be used with registered
/// `JsonSupportedCosmosMessages` (in case the contract is a DAO with **predefined actions**)
///
/// This enum does not derive [Deserialize](serde::Deserialize), see issue
/// [#1443](https://github.com/CosmWasm/cosmwasm/issues/1443)
#[derive(serde::Serialize, Clone, Debug, PartialEq, Eq)]
#[cfg_attr(test, derive(serde::Deserialize))]
#[serde(tag = "@type")]
enum JsonSupportedCosmosMessages {
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

impl ToString for JsonSupportedCosmosMessages {
    fn to_string(&self) -> String {
        serde_json_wasm::to_string(self).unwrap()
    }
}

mod msg_transfer {
    #[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq, Eq)]
    pub struct Height {
        pub revision_number: u64,
        pub revision_height: u64,
    }
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::{coins, from_json};

    use crate::ibc::types::packet::IcaPacketData;

    use super::JsonSupportedCosmosMessages;

    #[test]
    fn test_json_support() {
        #[derive(serde::Serialize, serde::Deserialize)]
        struct TestCosmosTx {
            pub messages: Vec<JsonSupportedCosmosMessages>,
        }

        let packet_from_string = IcaPacketData::from_json_strings(
            &[r#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#.to_string()], None);

        let packet_data = packet_from_string.data;
        let cosmos_tx: TestCosmosTx = from_json(packet_data).unwrap();

        let expected = JsonSupportedCosmosMessages::MsgSend {
            from_address: "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk".to_string(),
            to_address: "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk".to_string(),
            amount: coins(5000, "stake".to_string()),
        };

        assert_eq!(expected, cosmos_tx.messages[0]);
    }
}
