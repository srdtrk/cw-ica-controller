//! This module contains the helpers to convert [`QueryRequest`] to protobuf bytes and vice versa.

use cosmwasm_std::{Empty, QueryRequest};

pub use response::*;

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
        /// Response for a [`cosmwasm_std::StakingQuery`].
        #[cfg(feature = "staking")]
        Staking(StakingQueryResponse),
    }

    /// The response type for the [`cosmwasm_std::BankQuery`] queries.
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

    /// The response type for the [`cosmwasm_std::StakingQuery`] queries.
    #[cfg(feature = "staking")]
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
    #[cfg(feature = "staking")]
    #[cw_serde]
    pub struct IcaDelegationResponse {
        /// The delegation response if it exists.
        pub delegation: Option<Delegation>,
    }

    /// Response for the [`cosmwasm_std::StakingQuery::AllDelegations`] query over ICA.
    #[cfg(feature = "staking")]
    #[cw_serde]
    pub struct IcaAllDelegationsResponse {
        /// The delegations.
        pub delegations: Vec<Delegation>,
    }

    /// Delegation is the detailed information about a delegation.
    #[cfg(feature = "staking")]
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

/// Converts a [`QueryRequest`] to a gRPC method path and protobuf bytes.
///
/// # Panics
///
/// Panics if the query type is not supported.
#[must_use]
pub fn query_to_protobuf(query: QueryRequest<Empty>) -> (String, Vec<u8>) {
    match query {
        QueryRequest::Bank(bank_query) => convert_to_protobuf::bank(bank_query),
        QueryRequest::Stargate { path, data } => (path, data.0),
        QueryRequest::Wasm(_) => panic!("wasmd queries are not marked module safe (yet)"),
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

mod convert_to_protobuf {
    use cosmos_sdk_proto::{
        cosmos::bank::v1beta1::{
            QueryAllBalancesRequest, QueryBalanceRequest, QueryDenomMetadataRequest,
            QueryDenomsMetadataRequest,
        },
        cosmos::{bank::v1beta1::QuerySupplyOfRequest, base::query::v1beta1::PageRequest},
        prost::Message,
    };
    use cosmwasm_std::BankQuery;

    pub fn bank(bank_query: BankQuery) -> (String, Vec<u8>) {
        match bank_query {
            BankQuery::Balance { address, denom } => (
                "/cosmos.bank.v1beta1.Query/Balance".to_string(),
                QueryBalanceRequest { address, denom }.encode_to_vec(),
            ),
            BankQuery::AllBalances { address } => (
                "/cosmos.bank.v1beta1.Query/AllBalances".to_string(),
                QueryAllBalancesRequest {
                    address,
                    pagination: None,
                }
                .encode_to_vec(),
            ),
            BankQuery::DenomMetadata { denom } => (
                "/cosmos.bank.v1beta1.Query/DenomMetadata".to_string(),
                QueryDenomMetadataRequest { denom }.encode_to_vec(),
            ),
            BankQuery::AllDenomMetadata { pagination } => {
                let pagination = pagination.map(|pagination| PageRequest {
                    key: pagination.key.unwrap_or_default().0,
                    limit: u64::from(pagination.limit),
                    reverse: pagination.reverse,
                    offset: 0,
                    count_total: false,
                });

                (
                    "/cosmos.bank.v1beta1.Query/DenomsMetadata".to_string(),
                    QueryDenomsMetadataRequest { pagination }.encode_to_vec(),
                )
            }
            BankQuery::Supply { denom } => (
                "/cosmos.bank.v1beta1.Query/SupplyOf".to_string(),
                QuerySupplyOfRequest { denom }.encode_to_vec(),
            ),
            _ => panic!("Unsupported BankQuery"),
        }
    }

    #[cfg(feature = "staking")]
    pub fn staking(staking_query: cosmwasm_std::StakingQuery) -> (String, Vec<u8>) {
        use cosmos_sdk_proto::cosmos::staking::v1beta1::{
            QueryDelegationRequest, QueryDelegatorDelegationsRequest, QueryParamsRequest,
            QueryValidatorRequest, QueryValidatorsRequest,
        };

        match staking_query {
            cosmwasm_std::StakingQuery::Validator { address } => (
                "/cosmos.staking.v1beta1.Query/Validator".to_string(),
                QueryValidatorRequest {
                    validator_addr: address,
                }
                .encode_to_vec(),
            ),
            cosmwasm_std::StakingQuery::AllValidators {} => (
                "/cosmos.staking.v1beta1.Query/Validators".to_string(),
                QueryValidatorsRequest {
                    status: String::default(),
                    pagination: None,
                }
                .encode_to_vec(),
            ),
            cosmwasm_std::StakingQuery::Delegation {
                delegator,
                validator,
            } => (
                "/cosmos.staking.v1beta1.Query/Delegation".to_string(),
                QueryDelegationRequest {
                    delegator_addr: delegator,
                    validator_addr: validator,
                }
                .encode_to_vec(),
            ),
            cosmwasm_std::StakingQuery::AllDelegations { delegator } => (
                "/cosmos.staking.v1beta1.Query/DelegatorDelegations".to_string(),
                QueryDelegatorDelegationsRequest {
                    delegator_addr: delegator,
                    pagination: None,
                }
                .encode_to_vec(),
            ),
            cosmwasm_std::StakingQuery::BondedDenom {} => (
                "/cosmos.staking.v1beta1.Query/Params".to_string(),
                QueryParamsRequest::default().encode_to_vec(),
            ),
            _ => panic!("Unsupported StakingQuery"),
        }
    }
}

/// TODO
pub mod from_protobuf {
    use std::str::FromStr;

    use super::{BankQueryResponse, IcaQueryResponse};

    use crate::types::ContractError;

    use cosmos_sdk_proto::{
        cosmos::{
            bank::v1beta1::{
                Metadata as ProtoMetadata, QueryAllBalancesResponse, QueryBalanceResponse,
                QueryDenomMetadataResponse, QueryDenomsMetadataResponse, QuerySupplyOfResponse,
            },
            base::v1beta1::Coin as ProtoCoin,
        },
        prost::Message,
    };
    use cosmwasm_std::{
        AllBalanceResponse, AllDenomMetadataResponse, BalanceResponse, Binary, Coin, DenomMetadata,
        DenomMetadataResponse, DenomUnit, StdResult, SupplyResponse, Uint128,
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

        Ok(cosmwasm_std::Validator {
            address: validator.operator_address,
            commission: Decimal::from_str(&commission_rates.rate)?,
            max_commission: Decimal::from_str(&commission_rates.max_rate)?,
            max_change_rate: Decimal::from_str(&commission_rates.max_change_rate)?,
        })
    }

    /// Converts the response bytes to a [`IcaQueryResponse`] using the query path.
    ///
    /// # Errors
    /// Returns an error if the response bytes cannot be decoded.
    pub fn response(
        path: &str,
        resp: Vec<u8>,
        is_stargate: bool,
    ) -> Result<IcaQueryResponse, ContractError> {
        if is_stargate {
            return Ok(IcaQueryResponse::Stargate {
                data: Binary(resp),
                path: path.to_string(),
            });
        }

        match path {
            x if x.starts_with("/cosmos.bank.v1beta1.Query/") => bank_response(path, resp.as_ref()),
            #[cfg(feature = "staking")]
            x if x.starts_with("/cosmos.staking.v1beta1.Query/") => {
                staking_response(path, resp.as_ref())
            }
            _ => Err(ContractError::UnknownDataType(path.to_string())),
        }
    }

    fn bank_response(path: &str, resp: &[u8]) -> Result<IcaQueryResponse, ContractError> {
        match path {
            "/cosmos.bank.v1beta1.Query/Balance" => {
                let resp = QueryBalanceResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::Balance(
                    BalanceResponse::new(
                        resp.balance
                            .map_or_else(|| Ok(Coin::default()), convert_to_coin)?,
                    ),
                )))
            }
            "/cosmos.bank.v1beta1.Query/AllBalances" => {
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
            "/cosmos.bank.v1beta1.Query/DenomMetadata" => {
                let resp = QueryDenomMetadataResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::DenomMetadata(
                    DenomMetadataResponse::new(
                        resp.metadata
                            .map_or_else(DenomMetadata::default, convert_to_metadata),
                    ),
                )))
            }
            "/cosmos.bank.v1beta1.Query/DenomsMetadata" => {
                let resp = QueryDenomsMetadataResponse::decode(resp)?;
                Ok(IcaQueryResponse::Bank(BankQueryResponse::AllDenomMetadata(
                    AllDenomMetadataResponse::new(
                        resp.metadatas
                            .into_iter()
                            .map(convert_to_metadata)
                            .collect(),
                        resp.pagination
                            .map(|pagination| Binary(pagination.next_key)),
                    ),
                )))
            }
            "/cosmos.bank.v1beta1.Query/SupplyOf" => {
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
            "/cosmos.staking.v1beta1.Query/Validator" => {
                let resp = QueryValidatorResponse::decode(resp)?;
                Ok(IcaQueryResponse::Staking(StakingQueryResponse::Validator(
                    ValidatorResponse::new(resp.validator.map(convert_to_validator).transpose()?),
                )))
            }
            "/cosmos.staking.v1beta1.Query/Validators" => {
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
            "/cosmos.staking.v1beta1.Query/Delegation" => {
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
            "/cosmos.staking.v1beta1.Query/DelegatorDelegations" => {
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
            "/cosmos.staking.v1beta1.Query/Params" => {
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
        /// `path` defines the path of the query request as defined by ADR-021.
        /// https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-021-protobuf-query-encoding.md#custom-query-registration-and-routing
        pub path: ::prost::alloc::string::String,
        #[prost(bytes = "vec", tag = "2")]
        /// `data` defines the payload of the query request as defined by ADR-021.
        /// https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-021-protobuf-query-encoding.md#custom-query-registration-and-routing
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
}
