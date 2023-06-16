use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("unauthorized")]
    Unauthorized {},

    #[error("invalid channel ordering")]
    InvalidChannelOrdering {},

    #[error("invalid host port")]
    InvalidHostPort {},

    #[error("invalid controller port")]
    InvalidControllerPort {},

    #[error("invalid interchain accounts version: expected {expected}, got {actual}")]
    InvalidVersion { expected: String, actual: String },

    #[error("codec is not supported: unsupported codec format {0}")]
    UnsupportedCodec(String),

    #[error("invalid account address")]
    InvalidAddress {},

    #[error("unsupported transaction type {0}")]
    UnsupportedTxType(String),

    #[error("invalid connection")]
    InvalidConnection {},

    #[error("unknown data type: {0}")]
    UnknownDataType(String),

    #[error("active channel already set for this contract")]
    ActiveChannelAlreadySet {},

    #[error("invalid channel in contract state")]
    InvalidChannelInContractState {},
}
