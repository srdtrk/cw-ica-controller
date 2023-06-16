use cosmwasm_std::{IbcOrder, StdError};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("invalid channel ordering")]
    InvalidChannelOrdering {},

    #[error("invalid host port")]
    InvalidHostPort {},

    #[error("invalid interchain accounts version: expected {expected}, got {actual}")]
    InvalidVersion { expected: String, actual: String },
}
