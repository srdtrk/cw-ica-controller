//! # Metadata
//!
//! This file contains the [`IcaMetadata`] struct and its methods.
//!
//! The metadata is the information that is communicated during the handshake between the
//! ICA controller and the ICA host. It encodes key information about the messages exchanged
//! between the ICA controller and the ICA host.

use cosmwasm_std::IbcChannel;
use serde::{Deserialize, Serialize};

use crate::types::ContractError;

use super::keys::ICA_VERSION;

/// IcaMetadata is the metadata of the IBC application communicated during the handshake.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct IcaMetadata {
    /// The version of the IBC application.
    pub version: String,
    /// Controller's connection id.
    pub controller_connection_id: String,
    /// Counterparty's connection id.
    pub host_connection_id: String,
    /// The address of the interchain account.
    /// This address can be left empty at the `OpenInit` stage,
    /// and the ICA host will fill it in later during the handshake.
    pub address: String,
    /// The encoding of the messages sent to the ICA host.
    /// This contract only supports json encoding.
    pub encoding: String,
    /// The type of transaction that is sent to the ICA host.
    /// There is currently only one supported type: `sdk_multi_msg`.
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
            // counterparty connection_id is not exposed in the IbcChannel struct
            // so handshake will fail in non-test environments if this function is used
            host_connection_id: channel.connection_id.clone(),
            address: "".to_string(),
            encoding: "proto3json".to_string(),
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
        // We cannot check the counterparty connection_id because it is not exposed to the contract
        // if self.host_connection_id != channel.counterparty_endpoint.connection_id {
        //     return Err(ContractError::InvalidConnection {});
        // }
        if !self.address.is_empty() {
            validate_ica_address(&self.address)?;
        }
        if self.encoding != "proto3json" {
            return Err(ContractError::UnsupportedCodec(self.encoding.clone()));
        }
        if self.tx_type != "sdk_multi_msg" {
            return Err(ContractError::UnsupportedTxType(self.tx_type.clone()));
        }
        Ok(())
    }

    /// Checks if the previous version of the IcaMetadata is equal to the current one
    pub fn is_previous_version_equal(&self, previous_version: impl Into<String>) -> bool {
        let maybe_previous_metadata: Result<Self, _> =
            serde_json_wasm::from_str(&previous_version.into());
        match maybe_previous_metadata {
            Ok(previous_metadata) => {
                self.version == previous_metadata.version
                    && self.controller_connection_id == previous_metadata.controller_connection_id
                    && self.host_connection_id == previous_metadata.host_connection_id
                    && self.encoding == previous_metadata.encoding
                    && self.tx_type == previous_metadata.tx_type
            }
            Err(_) => false,
        }
    }
}

impl ToString for IcaMetadata {
    fn to_string(&self) -> String {
        serde_json_wasm::to_string(self).unwrap()
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

#[cfg(test)]
mod tests {
    use cosmwasm_std::{IbcEndpoint, IbcOrder};

    use super::*;

    fn mock_channel(
        version: impl Into<String>,
        connection_id: impl Into<String>,
        channel_id: impl Into<String>,
        port_id: impl Into<String>,
        counterparty_channel_id: impl Into<String>,
        counterparty_port_id: impl Into<String>,
    ) -> IbcChannel {
        let mock_endpoint = IbcEndpoint {
            port_id: port_id.into(),
            channel_id: channel_id.into(),
        };
        let mock_counterparty_endpoint = IbcEndpoint {
            port_id: counterparty_port_id.into(),
            channel_id: counterparty_channel_id.into(),
        };

        IbcChannel::new(
            mock_endpoint,
            mock_counterparty_endpoint,
            IbcOrder::Ordered,
            version,
            connection_id,
        )
    }

    fn mock_metadata() -> IcaMetadata {
        IcaMetadata::new(
            ICA_VERSION.to_string(),
            "connection-0".to_string(),
            "connection-1".to_string(),
            "".to_string(),
            "proto3json".to_string(),
            "sdk_multi_msg".to_string(),
        )
    }

    #[test]
    fn test_validate_success() {
        let channel = mock_channel(
            "ics27-1",
            "connection-0",
            "channel-0",
            "port-0",
            "channel-1",
            "port-1",
        );
        let metadata = IcaMetadata::from_channel(&channel);
        assert!(metadata.validate(&channel).is_ok());
    }

    #[test]
    fn test_validate_fail() {
        let channel_1 = mock_channel(
            "ics27-1",
            "connection-0",
            "channel-0",
            "port-0",
            "channel-1",
            "port-1",
        );
        let channel_2 = mock_channel(
            "ics27-1",
            "connection-1",
            "channel-0",
            "port-0",
            "channel-1",
            "port-1",
        );
        let metadata = IcaMetadata::from_channel(&channel_1);
        assert!(metadata.validate(&channel_2).is_err());
    }

    #[test]
    fn test_to_string() {
        let metadata = mock_metadata();
        let serialized_metadata = metadata.to_string();
        assert_eq!(
            serialized_metadata,
            r#"{"version":"ics27-1","controller_connection_id":"connection-0","host_connection_id":"connection-1","address":"","encoding":"proto3json","tx_type":"sdk_multi_msg"}"#
        );
    }

    #[test]
    fn test_deserialize_str() {
        let serialized_metadata = r#"{"version":"ics27-1","controller_connection_id":"connection-0","host_connection_id":"connection-1","address":"","encoding":"proto3json","tx_type":"sdk_multi_msg"}"#;
        let metadata: IcaMetadata = serde_json_wasm::from_str(serialized_metadata).unwrap();
        assert_eq!(metadata, mock_metadata());
    }

    #[test]
    fn test_is_previous_version_equal_success() {
        let metadata = mock_metadata();
        let previous_version = r#"{"version":"ics27-1","controller_connection_id":"connection-0","host_connection_id":"connection-1","address":"different","encoding":"proto3json","tx_type":"sdk_multi_msg"}"#;
        assert!(metadata.is_previous_version_equal(previous_version));
    }

    #[test]
    fn test_is_previous_version_equal_failure() {
        let metadata = mock_metadata();
        let previous_version = r#"{"version":"ics27-2","controller_connection_id":"connection-123","host_connection_id":"connection-11","address":"different","encoding":"proto3json","tx_type":"sdk_multi_msg"}"#;
        assert!(!metadata.is_previous_version_equal(previous_version));
    }
}
