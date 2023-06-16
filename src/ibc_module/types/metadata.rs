use cosmwasm_std::IbcChannel;
use serde::{Deserialize, Serialize};

use crate::ContractError;

use super::keys::ICA_VERSION;

/// IcaMetadata is the metadata of the IBC application communicated during the handshake
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct IcaMetadata {
    pub version: String,
    pub controller_connection_id: String,
    pub host_connection_id: String,
    pub address: String,
    pub encoding: String,
    pub tx_type: String,
}

impl IcaMetadata {
    /// Creates a new IcaMetadata
    pub fn new(
        version: String,
        controller_connection_id: String,
        host_connection_id: String,
        address: String,
        encoding: String,
        tx_type: String,
    ) -> Self {
        Self {
            version,
            controller_connection_id,
            host_connection_id,
            address,
            encoding,
            tx_type,
        }
    }
    /// Creates a new IcaMetadata from an IbcChannel
    pub fn from_channel(channel: &IbcChannel) -> Self {
        Self {
            version: ICA_VERSION.to_string(),
            controller_connection_id: channel.connection_id.clone(),
            host_connection_id: channel.counterparty_endpoint.channel_id.clone(),
            address: "".to_string(),
            encoding: "json".to_string(),
            tx_type: "sdk_multi_msg".to_string(),
        }
    }

    /// Validates the IcaMetadata
    pub fn validate(&self, channel: &IbcChannel) -> Result<(), ContractError> {
        if self.version != ICA_VERSION {
            return Err(ContractError::InvalidVersion {
                expected: ICA_VERSION.to_string(),
                actual: self.version.clone(),
            });
        }
        if self.controller_connection_id != channel.connection_id {
            return Err(ContractError::InvalidConnection {});
        }
        if self.host_connection_id != channel.counterparty_endpoint.channel_id {
            return Err(ContractError::InvalidConnection {});
        }
        if !self.address.is_empty() {
            validate_ica_address(&self.address)?;
        }
        if self.encoding != "json" {
            return Err(ContractError::UnsupportedCodec(self.encoding.clone()));
        }
        if self.tx_type != "sdk_multi_msg" {
            return Err(ContractError::UnsupportedTxType(self.tx_type.clone()));
        }
        Ok(())
    }
}

impl ToString for IcaMetadata {
    fn to_string(&self) -> String {
        serde_json::to_string(self).unwrap()
    }
}

/// Validates an ICA address
fn validate_ica_address(address: &str) -> Result<(), ContractError> {
    const DEFAULT_MAX_LENGTH: usize = 128;
    if address.len() > DEFAULT_MAX_LENGTH || !address.chars().all(|c| c.is_alphanumeric()) {
        return Err(ContractError::InvalidAddress {});
    }
    Ok(())
}
