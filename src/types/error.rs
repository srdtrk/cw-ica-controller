//! This module defines [`ContractError`].

use std::string::FromUtf8Error;

use cosmwasm_std::StdError;
use thiserror::Error;

/// `ContractError` is the error type returned by contract's functions.
#[allow(missing_docs)]
#[allow(clippy::module_name_repetitions)]
#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("FromUtf8Error: {0}")]
    JsonSerde(#[from] FromUtf8Error),

    #[error("json_serde_wasm serialization error: {0}")]
    JsonWasmSerialize(#[from] serde_json_wasm::ser::Error),

    #[error("json_serde_wasm deserialization error: {0}")]
    JsonWasmDeserialize(#[from] serde_json_wasm::de::Error),

    #[error("prost encoding error: {0}")]
    ProstEncodeError(#[from] cosmos_sdk_proto::prost::EncodeError),

    #[error("semver parse error: {0}")]
    SemverError(#[from] semver::Error),

    #[error("{0}")]
    OwnershipError(#[from] cw_ownable::OwnershipError),

    #[error("this contract must have an owner")]
    OwnershipCannotBeRenounced,

    #[error("unauthorized")]
    Unauthorized,

    #[error("invalid migration version: expected {expected}, got {actual}")]
    InvalidMigrationVersion { expected: String, actual: String },

    #[error("invalid channel ordering")]
    InvalidChannelOrdering,

    #[error("invalid host port")]
    InvalidHostPort,

    #[error("invalid controller port")]
    InvalidControllerPort,

    #[error("invalid interchain accounts version: expected {expected}, got {actual}")]
    InvalidVersion { expected: String, actual: String },

    #[error("MsgChannelOpenInit is not allowed")]
    ChannelOpenInitNotAllowed,

    #[error("MsgChannelCloseInit is not allowed")]
    ChannelCloseInitNotAllowed,

    #[error("codec is not supported: unsupported codec format {0}")]
    UnsupportedCodec(String),

    #[error("invalid interchain account address")]
    InvalidIcaAddress,

    #[error("unsupported transaction type {0}")]
    UnsupportedTxType(String),

    #[error("invalid connection")]
    InvalidConnection,

    #[error("unknown data type: {0}")]
    UnknownDataType(String),

    #[error("active channel already set for this contract")]
    ActiveChannelAlreadySet,

    #[error("invalid channel in contract state")]
    InvalidChannelInContractState,

    #[error("interchain account information is not set")]
    IcaInfoNotSet,

    #[error("no channel init options are provided to the contract")]
    NoChannelInitOptions,

    #[error("invalid channel status: expected {expected}, got {actual}")]
    InvalidChannelStatus { expected: String, actual: String },

    #[error("unsupported packet encoding: {0}")]
    UnsupportedPacketEncoding(String),
}
