//! This module contains the helpers to convert [`QueryRequest`] to protobuf bytes and vice versa.

use cosmwasm_std::{Empty, QueryRequest};

pub use response::*;

/// Converts a [`QueryRequest`] to a grpc method path, protobuf bytes, and a flag indicating if the query is stargate.
///
/// # Panics
///
/// Panics if the query type is not supported.
#[must_use]
pub fn query_to_protobuf(query: QueryRequest<Empty>) -> (String, Vec<u8>, bool) {
    match query {
        QueryRequest::Bank(bank_query) => convert_to_protobuf::bank(bank_query),
        #[allow(deprecated)]
        QueryRequest::Stargate { path, data } => (path, data.into(), true),
        QueryRequest::Wasm(wasm_query) => convert_to_protobuf::wasm(wasm_query),
        QueryRequest::Ibc(_) => panic!("ibc-go queries are not marked module safe (yet)"),
        QueryRequest::Custom(_) => panic!("custom queries are not supported"),
        #[cfg(feature = "staking")]
        QueryRequest::Staking(staking_query) => convert_to_protobuf::staking(staking_query),
        #[cfg(feature = "staking")]
        QueryRequest::Distribution(_) => {
            panic!("distribution queries are not marked module safe (yet)")
        }
        _ => panic!("Unsupported QueryRequest"),
    }
}

/// Converts [`proto::MsgModuleQuerySafeResponse`] to [`IcaQueryResult`] using the storage.
#[cfg(feature = "export")]
#[must_use]
pub fn result_from_response(
    paths: Vec<(String, bool)>,
    resp_msg: &proto::MsgModuleQuerySafeResponse,
) -> IcaQueryResult {
    if paths.len() != resp_msg.responses.len() {
        return IcaQueryResult::Error(format!(
            "expected {} responses, got {}",
            paths.len(),
            resp_msg.responses.len()
        ));
    }

    paths
        .into_iter()
        .zip(resp_msg.responses.iter())
        .map(|((path, is_stargate), resp)| from_protobuf::response(&path, resp, is_stargate))
        .collect::<Result<_, _>>()
        .map_or_else(
            |e| IcaQueryResult::Error(e.to_string()),
            |responses| IcaQueryResult::Success {
                height: resp_msg.height,
                responses,
            },
        )
}

/// The constants for the `query_msg` module.
pub mod constants {
    /// The query path for the Balance query.
    pub const BALANCE: &str = "/cosmos.bank.v1beta1.Query/Balance";
    /// The query path for the `AllBalances` query.
    pub const ALL_BALANCES: &str = "/cosmos.bank.v1beta1.Query/AllBalances";
    /// The query path for the `DenomMetadata` query.
    pub const DENOM_METADATA: &str = "/cosmos.bank.v1beta1.Query/DenomMetadata";
    /// The query path for the `AllDenomMetadata` query.
    pub const ALL_DENOM_METADATA: &str = "/cosmos.bank.v1beta1.Query/DenomsMetadata";
    /// The query path for the Supply query.
    pub const SUPPLY: &str = "/cosmos.bank.v1beta1.Query/SupplyOf";

    /// The query path for the Validator query.
    #[cfg(feature = "staking")]
    pub const VALIDATOR: &str = "/cosmos.staking.v1beta1.Query/Validator";
    /// The query path for the `AllValidators` query.
    #[cfg(feature = "staking")]
    pub const ALL_VALIDATORS: &str = "/cosmos.staking.v1beta1.Query/Validators";
    /// The query path for the Delegation query.
    #[cfg(feature = "staking")]
    pub const DELEGATION: &str = "/cosmos.staking.v1beta1.Query/Delegation";
    /// The query path for the `AllDelegations` query.
    #[cfg(feature = "staking")]
    pub const ALL_DELEGATIONS: &str = "/cosmos.staking.v1beta1.Query/DelegatorDelegations";
    /// The query path for the `BondedDenom` query.
    #[cfg(feature = "staking")]
    pub const STAKING_PARAMS: &str = "/cosmos.staking.v1beta1.Query/Params";

    /// The query path for the `ContractInfo` query.
    pub const WASM_CONTRACT_INFO: &str = "/cosmwasm.wasm.v1.Query/ContractInfo";
    /// The query path for the `CodeInfo` query.
    pub const WASM_CODE: &str = "/cosmwasm.wasm.v1.Query/Code";
    /// The query path for the Raw query.
    pub const WASM_RAW: &str = "/cosmwasm.wasm.v1.Query/RawContractState";
    /// The query path for the Smart query.
    pub const WASM_SMART: &str = "/cosmwasm.wasm.v1.Query/SmartContractState";
}

#[allow(clippy::module_name_repetitions)]
mod response {
    use cosmwasm_schema::cw_serde;

    /// The result of an ICA query packet.
    #[cw_serde]
    pub enum IcaQueryResult {
        /// The query was successful and the responses are included.
        Success {
            /// The height of the block at which the queries were executed on the counterparty chain.
            height: u64,
            /// The responses to the queries.
            responses: Vec<IcaQueryResponse>,
        },
        /// The query failed with an error message. The error string
        /// often does not contain useful information for the end user.
        Error(String),
    }

    /// The response for a successful ICA query.
    #[non_exhaustive]
    #[cw_serde]
    pub enum IcaQueryResponse {
        /// Response for a [`cosmwasm_std::BankQuery`].
        Bank(BankQueryResponse),
        /// Response for a [`cosmwasm_std::QueryRequest::Stargate`].
        /// Protobuf encoded bytes stored as [`cosmwasm_std::Binary`].
        Stargate {
            /// The response bytes.
            data: cosmwasm_std::Binary,
            /// The query grpc method
            path: String,
        },
        /// Response for a [`cosmwasm_std::WasmQuery`].
        Wasm(WasmQueryResponse),
        /// Response for a [`cosmwasm_std::StakingQuery`].
        Staking(StakingQueryResponse),
    }

    /// The response type for the [`cosmwasm_std::BankQuery`] queries.
    #[non_exhaustive]
    #[cw_serde]
    pub enum BankQueryResponse {
        /// Response for the [`cosmwasm_std::BankQuery::Supply`] query.
        Supply(cosmwasm_std::SupplyResponse),
        /// Response for the [`cosmwasm_std::BankQuery::Balance`] query.
        Balance(cosmwasm_std::BalanceResponse),
        /// Response for the [`cosmwasm_std::BankQuery::AllBalances`] query.
        AllBalances(cosmwasm_std::AllBalanceResponse),
        /// Response for the [`cosmwasm_std::BankQuery::DenomMetadata`] query.
        DenomMetadata(cosmwasm_std::DenomMetadataResponse),
        /// Response for the [`cosmwasm_std::BankQuery::AllDenomMetadata`] query.
        AllDenomMetadata(cosmwasm_std::AllDenomMetadataResponse),
    }

    /// The response type for the [`cosmwasm_std::WasmQuery`] queries.
    #[non_exhaustive]
    #[cw_serde]
    pub enum WasmQueryResponse {
        /// Response for the [`cosmwasm_std::WasmQuery::ContractInfo`] query.
        /// Returns `None` if the contract does not exist.
        /// The `pinned` field is not supported.
        ContractInfo(Option<cosmwasm_std::ContractInfoResponse>),
        /// Response for the [`cosmwasm_std::WasmQuery::CodeInfo`] query.
        /// Returns `None` if the code does not exist.
        CodeInfo(Option<cosmwasm_std::CodeInfoResponse>),
        /// Response for the [`cosmwasm_std::WasmQuery::Raw`] query.
        RawContractState(Option<cosmwasm_std::Binary>),
        /// Response for the [`cosmwasm_std::WasmQuery::Smart`] query.
        SmartContractState(cosmwasm_std::Binary),
    }

    /// The response type for the [`cosmwasm_std::StakingQuery`] queries.
    #[non_exhaustive]
    #[cw_serde]
    pub enum StakingQueryResponse {
        /// Response for the [`cosmwasm_std::StakingQuery::BondedDenom`] query.
        BondedDenom(cosmwasm_std::BondedDenomResponse),
        /// Response for the [`cosmwasm_std::StakingQuery::AllDelegations`] query.
        AllDelegations(IcaAllDelegationsResponse),
        /// Response for the [`cosmwasm_std::StakingQuery::Delegation`] query.
        Delegation(IcaDelegationResponse),
        /// Response for the [`cosmwasm_std::StakingQuery::AllValidators`] query.
        AllValidators(cosmwasm_std::AllValidatorsResponse),
        /// Response for the [`cosmwasm_std::StakingQuery::Validator`] query.
        Validator(cosmwasm_std::ValidatorResponse),
    }

    /// Response for the [`cosmwasm_std::StakingQuery::Delegation`] query over ICA.
    #[cw_serde]
    pub struct IcaDelegationResponse {
        /// The delegation response if it exists.
        pub delegation: Option<Delegation>,
    }

    /// Response for the [`cosmwasm_std::StakingQuery::AllDelegations`] query over ICA.
    #[cw_serde]
    pub struct IcaAllDelegationsResponse {
        /// The delegations.
        pub delegations: Vec<Delegation>,
    }

    /// Delegation is the detailed information about a delegation.
    #[cw_serde]
    pub struct Delegation {
        /// The delegator address.
        pub delegator: String,
        /// A validator address (e.g. cosmosvaloper1...)
        pub validator: String,
        /// Delegation amount.
        pub amount: cosmwasm_std::Coin,
    }
}

mod convert_to_protobuf {
    use cosmos_sdk_proto::{
        cosmos::{
            bank::v1beta1::{
                QueryAllBalancesRequest, QueryBalanceRequest, QueryDenomMetadataRequest,
                QueryDenomsMetadataRequest, QuerySupplyOfRequest,
            },
            base::query::v1beta1::PageRequest,
        },
        cosmwasm::wasm::v1::{
            QueryContractInfoRequest, QueryRawContractStateRequest, QuerySmartContractStateRequest,
        },
        prost::Message,
    };

    use cosmwasm_std::{BankQuery, WasmQuery};

    use super::constants;

    pub fn bank(bank_query: BankQuery) -> (String, Vec<u8>, bool) {
        match bank_query {
            BankQuery::Balance { address, denom } => (
                constants::BALANCE.to_string(),
                QueryBalanceRequest { address, denom }.encode_to_vec(),
                false,
            ),
            BankQuery::AllBalances { address } => (
                constants::ALL_BALANCES.to_string(),
                QueryAllBalancesRequest {
                    address,
                    pagination: None,
                }
                .encode_to_vec(),
                false,
            ),
            BankQuery::DenomMetadata { denom } => (
                constants::DENOM_METADATA.to_string(),
                QueryDenomMetadataRequest { denom }.encode_to_vec(),
                false,
            ),
            BankQuery::AllDenomMetadata { pagination } => {
                let pagination = pagination.map(|pagination| PageRequest {
                    key: pagination.key.unwrap_or_default().into(),
                    limit: u64::from(pagination.limit),
                    reverse: pagination.reverse,
                    offset: 0,
                    count_total: false,
                });

                (
                    constants::ALL_DENOM_METADATA.to_string(),
                    QueryDenomsMetadataRequest { pagination }.encode_to_vec(),
                    false,
                )
            }
            BankQuery::Supply { denom } => (
                constants::SUPPLY.to_string(),
                QuerySupplyOfRequest { denom }.encode_to_vec(),
                false,
            ),
            _ => panic!("Unsupported BankQuery"),
        }
    }

    pub fn wasm(wasm_query: WasmQuery) -> (String, Vec<u8>, bool) {
        match wasm_query {
            WasmQuery::Raw { contract_addr, key } => (
                constants::WASM_RAW.to_string(),
                QueryRawContractStateRequest {
                    address: contract_addr,
                    query_data: key.into(),
                }
                .encode_to_vec(),
                false,
            ),
            WasmQuery::Smart { contract_addr, msg } => (
                constants::WASM_SMART.to_string(),
                QuerySmartContractStateRequest {
                    address: contract_addr,
                    query_data: msg.into(),
                }
                .encode_to_vec(),
                false,
            ),
            WasmQuery::ContractInfo { contract_addr } => (
                constants::WASM_CONTRACT_INFO.to_string(),
                QueryContractInfoRequest {
                    address: contract_addr,
                }
                .encode_to_vec(),
                false,
            ),
            WasmQuery::CodeInfo { .. } => {
                panic!("CodeInfo query is not supported due to response size.")
            }
            _ => panic!("Unsupported WasmQuery"),
        }
    }

    #[cfg(feature = "staking")]
    pub fn staking(staking_query: cosmwasm_std::StakingQuery) -> (String, Vec<u8>, bool) {
        use cosmos_sdk_proto::cosmos::staking::v1beta1::{
            QueryDelegationRequest, QueryDelegatorDelegationsRequest, QueryParamsRequest,
            QueryValidatorRequest, QueryValidatorsRequest,
        };

        match staking_query {
            cosmwasm_std::StakingQuery::Validator { address } => (
                constants::VALIDATOR.to_string(),
                QueryValidatorRequest {
                    validator_addr: address,
                }
                .encode_to_vec(),
                false,
            ),
            cosmwasm_std::StakingQuery::AllValidators {} => (
                constants::ALL_VALIDATORS.to_string(),
                QueryValidatorsRequest {
                    status: String::default(),
                    pagination: None,
                }
                .encode_to_vec(),
                false,
            ),
            cosmwasm_std::StakingQuery::Delegation {
                delegator,
                validator,
            } => (
                constants::DELEGATION.to_string(),
                QueryDelegationRequest {
                    delegator_addr: delegator,
                    validator_addr: validator,
                }
                .encode_to_vec(),
                false,
            ),
            cosmwasm_std::StakingQuery::AllDelegations { delegator } => (
                constants::ALL_DELEGATIONS.to_string(),
                QueryDelegatorDelegationsRequest {
                    delegator_addr: delegator,
                    pagination: None,
                }
                .encode_to_vec(),
                false,
            ),
            cosmwasm_std::StakingQuery::BondedDenom {} => (
                constants::STAKING_PARAMS.to_string(),
                QueryParamsRequest::default().encode_to_vec(),
                false,
            ),
            _ => panic!("Unsupported StakingQuery"),
        }
    }
}

/// Converts the response bytes to a [`IcaQueryResponse`] using the query path.
pub mod from_protobuf {
    use std::str::FromStr;

    use super::{constants, BankQueryResponse, IcaQueryResponse, WasmQueryResponse};

    use crate::types::ContractError;

    use cosmos_sdk_proto::{
        cosmos::{
            bank::v1beta1::{
                Metadata as ProtoMetadata, QueryAllBalancesResponse, QueryBalanceResponse,
                QueryDenomMetadataResponse, QueryDenomsMetadataResponse, QuerySupplyOfResponse,
            },
            base::v1beta1::Coin as ProtoCoin,
        },
        cosmwasm::wasm::v1::{
            QueryContractInfoResponse, QueryRawContractStateResponse,
            QuerySmartContractStateResponse,
        },
        prost::Message,
    };
    use cosmwasm_std::{
        AllBalanceResponse, AllDenomMetadataResponse, BalanceResponse, Binary, Coin,
        ContractInfoResponse, DenomMetadata, DenomMetadataResponse, DenomUnit, StdResult,
        SupplyResponse, Uint128,
    };

    fn convert_to_coin(coin: ProtoCoin) -> StdResult<Coin> {
        Ok(Coin {
            denom: coin.denom,
            amount: Uint128::from_str(&coin.amount)?,
        })
    }

    fn convert_to_metadata(metadata: ProtoMetadata) -> DenomMetadata {
        DenomMetadata {
            name: metadata.name,
            symbol: metadata.symbol,
            description: metadata.description,
            base: metadata.base,
            display: metadata.display,
            uri: metadata.uri,
            uri_hash: metadata.uri_hash,
            denom_units: metadata
                .denom_units
                .into_iter()
                .map(|unit| DenomUnit {
                    denom: unit.denom,
                    exponent: unit.exponent,
                    aliases: unit.aliases,
                })
                .collect(),
        }
    }

    #[cfg(feature = "staking")]
    fn convert_to_validator(
        validator: cosmos_sdk_proto::cosmos::staking::v1beta1::Validator,
    ) -> StdResult<cosmwasm_std::Validator> {
        use cosmwasm_std::Decimal;

        let commission_rates = validator
            .commission
            .unwrap_or_default()
            .commission_rates
            .unwrap_or_default();

        Ok(cosmwasm_std::Validator::create(
            validator.operator_address,
            Decimal::from_str(&commission_rates.rate)?,
            Decimal::from_str(&commission_rates.max_rate)?,
            Decimal::from_str(&commission_rates.max_change_rate)?,
        ))
    }

    /// Converts the response bytes to a [`IcaQueryResponse`] using the query path.
    ///
    /// # Errors
    /// Returns an error if the response bytes cannot be decoded.
    pub fn response(
        path: &str,
        resp: &[u8],
        is_stargate: bool,
    ) -> Result<IcaQueryResponse, ContractError> {
        if is_stargate {
            return Ok(IcaQueryResponse::Stargate {
                data: Binary::from(resp),
                path: path.to_string(),
            });
        }

        match path {
            x if x.starts_with("/cosmos.bank.v1beta1.Query/") => bank_response(path, resp),
            x if x.starts_with("/cosmwasm.wasm.v1.Query/") => wasm_response(path, resp),
            #[cfg(feature = "staking")]
            x if x.starts_with("/cosmos.staking.v1beta1.Query/") => staking_response(path, resp),
            _ => Err(ContractError::UnknownDataType(path.to_string())),
        }
    }

    fn bank_response(path: &str, resp: &[u8]) -> Result<IcaQueryResponse, ContractError> {
        match path {
            constants::BALANCE => {
                let resp = QueryBalanceResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::Balance(
                    BalanceResponse::new(
                        resp.balance
                            .map_or_else(|| Ok(Coin::default()), convert_to_coin)?,
                    ),
                )))
            }
            constants::ALL_BALANCES => {
                let resp = QueryAllBalancesResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::AllBalances(
                    AllBalanceResponse::new(
                        resp.balances
                            .into_iter()
                            .map(convert_to_coin)
                            .collect::<StdResult<_>>()?,
                    ),
                )))
            }
            constants::DENOM_METADATA => {
                let resp = QueryDenomMetadataResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::DenomMetadata(
                    DenomMetadataResponse::new(
                        resp.metadata
                            .map_or_else(DenomMetadata::default, convert_to_metadata),
                    ),
                )))
            }
            constants::ALL_DENOM_METADATA => {
                let resp = QueryDenomsMetadataResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::AllDenomMetadata(
                    AllDenomMetadataResponse::new(
                        resp.metadatas
                            .into_iter()
                            .map(convert_to_metadata)
                            .collect(),
                        resp.pagination
                            .map(|pagination| Binary::new(pagination.next_key)),
                    ),
                )))
            }
            constants::SUPPLY => {
                let resp = QuerySupplyOfResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::Supply(
                    SupplyResponse::new(
                        resp.amount
                            .map_or_else(|| Ok(Coin::default()), convert_to_coin)?,
                    ),
                )))
            }
            _ => Err(ContractError::UnknownDataType(path.to_string())),
        }
    }

    fn wasm_response(path: &str, resp: &[u8]) -> Result<IcaQueryResponse, ContractError> {
        match path {
            constants::WASM_CONTRACT_INFO => {
                let resp = QueryContractInfoResponse::decode(resp)?;
                Ok(IcaQueryResponse::Wasm(WasmQueryResponse::ContractInfo(
                    resp.contract_info.map(|info| {
                        ContractInfoResponse::new(
                            info.code_id,
                            cosmwasm_std::Addr::unchecked(info.creator),
                            if info.admin.is_empty() {
                                None
                            } else {
                                Some(cosmwasm_std::Addr::unchecked(info.admin))
                            },
                            false,
                            if info.ibc_port_id.is_empty() {
                                None
                            } else {
                                Some(info.ibc_port_id)
                            },
                        )
                    }),
                )))
            }
            constants::WASM_RAW => {
                let resp = QueryRawContractStateResponse::decode(resp)?;
                Ok(IcaQueryResponse::Wasm(WasmQueryResponse::RawContractState(
                    if resp.data.is_empty() {
                        None
                    } else {
                        Some(resp.data.into())
                    },
                )))
            }
            constants::WASM_SMART => {
                let resp = QuerySmartContractStateResponse::decode(resp)?;
                Ok(IcaQueryResponse::Wasm(
                    WasmQueryResponse::SmartContractState(resp.data.into()),
                ))
            }
            _ => Err(ContractError::UnknownDataType(path.to_string())),
        }
    }

    #[cfg(feature = "staking")]
    fn staking_response(path: &str, resp: &[u8]) -> Result<IcaQueryResponse, ContractError> {
        use super::{
            Delegation, IcaAllDelegationsResponse, IcaDelegationResponse, StakingQueryResponse,
        };

        use cosmos_sdk_proto::cosmos::staking::v1beta1::{
            QueryDelegationResponse, QueryDelegatorDelegationsResponse, QueryParamsResponse,
            QueryValidatorResponse, QueryValidatorsResponse,
        };
        use cosmwasm_std::{AllValidatorsResponse, BondedDenomResponse, ValidatorResponse};

        match path {
            constants::VALIDATOR => {
                let resp = QueryValidatorResponse::decode(resp)?;
                Ok(IcaQueryResponse::Staking(StakingQueryResponse::Validator(
                    ValidatorResponse::new(resp.validator.map(convert_to_validator).transpose()?),
                )))
            }
            constants::ALL_VALIDATORS => {
                let resp = QueryValidatorsResponse::decode(resp)?;
                Ok(IcaQueryResponse::Staking(
                    StakingQueryResponse::AllValidators(AllValidatorsResponse::new(
                        resp.validators
                            .into_iter()
                            .map(convert_to_validator)
                            .collect::<StdResult<_>>()?,
                    )),
                ))
            }
            constants::DELEGATION => {
                let resp = QueryDelegationResponse::decode(resp)?;
                Ok(IcaQueryResponse::Staking(StakingQueryResponse::Delegation(
                    IcaDelegationResponse {
                        delegation: resp
                            .delegation_response
                            .and_then(|del_resp| {
                                del_resp.delegation.map(|del| -> StdResult<_> {
                                    Ok(Delegation {
                                        delegator: del.delegator_address,
                                        validator: del.validator_address,
                                        amount: convert_to_coin(
                                            del_resp.balance.unwrap_or_default(),
                                        )?,
                                    })
                                })
                            })
                            .transpose()?,
                    },
                )))
            }
            constants::ALL_DELEGATIONS => {
                let resp = QueryDelegatorDelegationsResponse::decode(resp)?;
                Ok(IcaQueryResponse::Staking(
                    StakingQueryResponse::AllDelegations(IcaAllDelegationsResponse {
                        delegations: resp
                            .delegation_responses
                            .into_iter()
                            .filter_map(|del_resp| {
                                del_resp.delegation.map(|del| -> StdResult<_> {
                                    Ok(Delegation {
                                        delegator: del.delegator_address,
                                        validator: del.validator_address,
                                        amount: convert_to_coin(
                                            del_resp.balance.unwrap_or_default(),
                                        )?,
                                    })
                                })
                            })
                            .collect::<StdResult<_>>()?,
                    }),
                ))
            }
            constants::STAKING_PARAMS => {
                let resp = QueryParamsResponse::decode(resp)?;
                Ok(IcaQueryResponse::Staking(
                    StakingQueryResponse::BondedDenom(BondedDenomResponse::new(
                        resp.params
                            .ok_or_else(|| ContractError::EmptyResponse(path.to_string()))?
                            .bond_denom,
                    )),
                ))
            }
            _ => Err(ContractError::UnknownDataType(path.to_string())),
        }
    }
}

/// This module defines the protobuf messages for the query module.
/// This module can be removed once these types are included in `cosmos_sdk_proto` crate.
// TODO: Remove this module once the types are included in `cosmos_sdk_proto` crate.
pub mod proto {
    /// `MsgModuleQuerySafe` defines the query request tx added in ibc-go v8.2
    #[derive(::prost::Message)]
    pub struct MsgModuleQuerySafe {
        #[prost(string, tag = "1")]
        /// signer is the address of the account that signed the transaction
        pub signer: ::prost::alloc::string::String,
        /// requests is the list of query requests
        #[prost(message, repeated, tag = "2")]
        pub requests: ::prost::alloc::vec::Vec<AbciQueryRequest>,
    }

    /// `AbciQueryRequest` defines the parameters for a particular query request by an interchain account.
    #[derive(::prost::Message)]
    pub struct AbciQueryRequest {
        #[prost(string, tag = "1")]
        /// `path` defines the path of the query request as defined by
        /// [ADR-021](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-021-protobuf-query-encoding.md#custom-query-registration-and-routing).
        pub path: ::prost::alloc::string::String,
        #[prost(bytes = "vec", tag = "2")]
        /// `data` defines the payload of the query request as defined by ADR-021.
        /// [ADR-021](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-021-protobuf-query-encoding.md#custom-query-registration-and-routing).
        pub data: ::prost::alloc::vec::Vec<u8>,
    }

    /// `MsgModuleQuerySafeResponse` defines the response for Msg/ModuleQuerySafe
    #[derive(::prost::Message)]
    pub struct MsgModuleQuerySafeResponse {
        /// height is the block height at which the query was executed
        #[prost(uint64, tag = "1")]
        pub height: u64,
        /// responses is the list of query responses as bytes
        /// The responses are in the same order as the requests
        #[prost(bytes = "vec", repeated, tag = "2")]
        pub responses: ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
    }

    impl ::prost::Name for MsgModuleQuerySafe {
        const NAME: &'static str = "MsgModuleQuerySafe";
        const PACKAGE: &'static str = "ibc.applications.interchain_accounts.host.v1";
    }
}
