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
/// * `msg`         - Base64-encoded json message to deserialize and insert the ICA address into.
/// * `ica_address` - ICA address to insert into the message.
pub fn insert_ica_address(msg: Binary, ica_address: impl Into<String>) -> StdResult<String> {
    let decoded: serde_json::Value = from_binary(&msg)?;
    Ok(decoded
        .to_string()
        .replace(ICA_PLACEHOLDER, &ica_address.into()))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_insert_ica_address() {
        let msg = Binary::from(br#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "$ica_address", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#);
        let ica_address = "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk";

        let expected: serde_json::Value = serde_json_wasm::from_str(
            r#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#,
        ).unwrap();
        let actual: serde_json::Value =
            serde_json_wasm::from_str(&insert_ica_address(msg, ica_address).unwrap()).unwrap();
        assert_eq!(actual, expected);
    }
}
