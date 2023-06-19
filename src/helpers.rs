use cosmwasm_std::{from_binary, Binary, StdResult};

/// ICA_PLACEHOLDER is used in custom messages to indicate the ICA address.
/// It is replaced with the ICA address before the message is sent to the ICA host.
pub const ICA_PLACEHOLDER: &str = "$ica_address";

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
