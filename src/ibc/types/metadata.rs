//! # Metadata
//!
//! This file contains the [`IcaMetadata`] struct and its methods.
//!
//! The version metadata is the information that is communicated during the ICS-27 channel
//! handshake between this contract and the ICA host. It encodes key information about the
//! interchain account.

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Deps, IbcChannel};

use crate::types::{
    state::{CHANNEL_OPEN_INIT_OPTIONS, CHANNEL_STATE},
    ContractError,
};

use super::keys::ICA_VERSION;

/// `IcaMetadata` is the metadata of the IBC application communicated during the handshake.
#[allow(clippy::module_name_repetitions)]
#[cw_serde]
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
    pub encoding: TxEncoding,
    /// The type of transaction that is sent to the ICA host.
    /// There is currently only one supported type: `sdk_multi_msg`.
    pub tx_type: String,
}

/// `TxEncoding` is the encoding of the transactions sent to the ICA host.
#[cw_serde]
pub enum TxEncoding {
    /// `Protobuf` is the protobuf serialization of the CosmosSDK's Any.
    #[serde(rename = "proto3")]
    Protobuf,
    /// `Proto3Json` is the json serialization of the CosmosSDK's Any.
    #[serde(rename = "proto3json")]
    Proto3Json,
}

impl IcaMetadata {
    /// Creates a new [`IcaMetadata`]
    #[must_use]
    pub const fn new(
        version: String,
        controller_connection_id: String,
        host_connection_id: String,
        address: String,
        encoding: TxEncoding,
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

    /// Creates a new [`IcaMetadata`] from an [`IbcChannel`]
    ///
    /// This is a fallback option if the ICA controller is not provided with the
    /// handshake version metadata by the relayer. It first tries to load the
    /// previous version of the [`IcaMetadata`] from the store, and if it fails,
    /// it uses the [`CHANNEL_OPEN_INIT_OPTIONS`] to create a new [`IcaMetadata`].
    ///
    /// # Errors
    ///
    /// Returns an error if the previous version of the [`IcaMetadata`] cannot be loaded
    /// from the store, and no [`CHANNEL_OPEN_INIT_OPTIONS`] are set in the store.
    pub fn from_channel(deps: Deps, channel: &IbcChannel) -> Result<Self, ContractError> {
        // If the the counterparty chain is using the fee middleware, and the this chain is not,
        // and the previous handshake was initiated with an empty version string, then the
        // previous version in the contract's channel state will be wrapped by the fee middleware,
        // and the IcaMetadata will not be able to be deserialized.
        if let Ok(channel_state) = CHANNEL_STATE.load(deps.storage) {
            if let Ok(previous_metadata) = serde_json_wasm::from_str(&channel_state.channel.version)
            {
                return Ok(previous_metadata);
            }
        }

        let options = CHANNEL_OPEN_INIT_OPTIONS.load(deps.storage)?;
        if options.connection_id != channel.connection_id {
            return Err(ContractError::InvalidConnection);
        }

        Ok(Self {
            version: ICA_VERSION.to_string(),
            encoding: TxEncoding::Protobuf,
            controller_connection_id: options.connection_id,
            // counterparty connection_id is not exposed to the contract, so we
            // use a stargate query to get it. Stargate queries are not universally
            // supported, so this is a fallback option.
            host_connection_id: options.counterparty_connection_id,
            address: String::new(),
            tx_type: "sdk_multi_msg".to_string(),
        })
    }

    /// Validates the [`IcaMetadata`]
    ///
    /// # Errors
    ///
    /// Returns an error if the metadata is invalid.
    pub fn validate(&self, channel: &IbcChannel) -> Result<(), ContractError> {
        if self.version != ICA_VERSION {
            return Err(ContractError::InvalidVersion {
                expected: ICA_VERSION.to_string(),
                actual: self.version.clone(),
            });
        }
        if self.controller_connection_id != channel.connection_id {
            return Err(ContractError::InvalidConnection);
        }
        if !matches!(self.encoding, TxEncoding::Protobuf) {
            return Err(ContractError::UnsupportedPacketEncoding(
                self.encoding.to_string(),
            ));
        }
        // We cannot check the counterparty connection_id because it is not exposed to the contract
        if !self.address.is_empty() {
            validate_ica_address(&self.address)?;
        }
        if self.tx_type != "sdk_multi_msg" {
            return Err(ContractError::UnsupportedTxType(self.tx_type.clone()));
        }
        Ok(())
    }
}

impl ToString for IcaMetadata {
    fn to_string(&self) -> String {
        serde_json_wasm::to_string(self).unwrap()
    }
}

impl ToString for TxEncoding {
    fn to_string(&self) -> String {
        serde_json_wasm::to_string(self).unwrap()
    }
}

/// Validates an ICA address
///
/// # Errors
///
/// Returns an error if the address is too long or contains invalid characters.
fn validate_ica_address(address: &str) -> Result<(), ContractError> {
    const DEFAULT_MAX_LENGTH: usize = 128;
    if address.len() > DEFAULT_MAX_LENGTH || !address.chars().all(char::is_alphanumeric) {
        return Err(ContractError::InvalidIcaAddress);
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::{testing::mock_dependencies, IbcEndpoint, IbcOrder};

    use crate::types::msg::options::ChannelOpenInitOptions;

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
            String::new(),
            TxEncoding::Proto3Json,
            "sdk_multi_msg".to_string(),
        )
    }

    #[test]
    fn test_validate_success() {
        let mut deps = mock_dependencies();

        let channel = mock_channel(
            "ics27-1",
            "connection-0",
            "channel-0",
            "port-0",
            "channel-1",
            "port-1",
        );
        let stored_init_options = ChannelOpenInitOptions {
            connection_id: "connection-0".to_string(),
            counterparty_connection_id: "connection-1".to_string(),
            counterparty_port_id: Some(super::super::keys::HOST_PORT_ID.to_string()),
            channel_ordering: None,
        };

        CHANNEL_OPEN_INIT_OPTIONS
            .save(deps.as_mut().storage, &stored_init_options)
            .unwrap();

        let metadata = IcaMetadata::from_channel(deps.as_ref(), &channel).unwrap();

        assert!(metadata.validate(&channel).is_ok());
    }

    #[test]
    fn test_validate_fail() {
        let mut deps = mock_dependencies();

        let channel_1 = mock_channel(
            "ics27-1",
            "connection-0",
            "channel-0",
            "port-0",
            "channel-1",
            "port-1",
        );

        let stored_init_options = ChannelOpenInitOptions {
            connection_id: "connection-0".to_string(),
            counterparty_connection_id: "connection-1".to_string(),
            counterparty_port_id: Some(super::super::keys::HOST_PORT_ID.to_string()),
            channel_ordering: None,
        };

        CHANNEL_OPEN_INIT_OPTIONS
            .save(deps.as_mut().storage, &stored_init_options)
            .unwrap();

        let channel_2 = mock_channel(
            "ics27-1",
            "connection-1",
            "channel-0",
            "port-0",
            "channel-1",
            "port-1",
        );
        let metadata = IcaMetadata::from_channel(deps.as_ref(), &channel_1).unwrap();
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
}
