use cosmwasm_std::{Instantiate2AddressError, StdError};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("error when computing the instantiate2 address: {0}")]
    Instantiate2AddressError(#[from] Instantiate2AddressError),

    #[error("prost encoding error: {0}")]
    ProstEncodeError(#[from] cosmos_sdk_proto::prost::EncodeError),

    #[error("unauthorized")]
    Unauthorized {},

    #[error("ica information is not set")]
    IcaInfoNotSet {},
}
