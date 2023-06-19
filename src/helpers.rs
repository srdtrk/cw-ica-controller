use cosmwasm_std::{from_binary, Binary, StdResult};

use crate::types::keys::ICA_PLACEHOLDER;

/// Returns StdResult<String>
///
/// Decodes the given base64-encoded json message and inserts the ICA address into it.
///
/// # Arguments
///
/// * `msg` - Base64-encoded json message to deserialize and insert the ICA address into.
/// * `ica_address` - ICA address to insert into the message.
pub fn insert_ica_address(msg: Binary, ica_address: impl Into<String>) -> StdResult<String> {
    let decoded: String = from_binary(&msg)?;
    Ok(decoded.replace(ICA_PLACEHOLDER, &ica_address.into()))
}
