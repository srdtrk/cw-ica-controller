//! # Stargate
//!
//! This module contains all the IBC stargate types that are needed to communicate with the IBC
//! core module. The use of this module is optional, and it currently only needed if the ICA controller
//! is not provided with the [handshake version metadata](super::metadata::IcaMetadata) by the relayer.
//!
//! Not all blockchains support the stargate messages, it is therefore recommended to provide the
//! handshake version metadata to the ICA controller. See a full discussion of this topic
//! [here](https://github.com/cosmos/ibc-go/issues/3942).
//!
//! This module is not tested in the end-to-end tests as the default wasmd docker image does not support
//! stargate queries. It is tested anecdotally, so use it at your own risk.

use cosmos_sdk_proto::traits::Message;

use crate::types::ContractError;
use cosmwasm_std::Binary;

/// Contains the stargate channel lifecycle helper methods.
pub mod channel {
    use super::*;

    use cosmos_sdk_proto::ibc::core::channel::v1::{
        Channel, Counterparty, MsgChannelOpenInit, Order, State,
    };

    use cosmwasm_std::CosmosMsg;

    use super::super::{keys, metadata};

    /// Creates a new MsgChannelOpenInit for an ica channel with the given contract address.
    /// Also generates the handshake version.
    /// If the counterparty port id is not provided, [`keys::HOST_PORT_ID`] is used.
    /// If the tx encoding is not provided, [`metadata::TxEncoding::Protobuf`] is used.
    pub fn new_ica_channel_open_init_cosmos_msg(
        contract_address: impl Into<String> + Clone,
        connection_id: impl Into<String> + Clone,
        counterparty_port_id: Option<impl Into<String>>,
        counterparty_connection_id: impl Into<String>,
        tx_encoding: Option<metadata::TxEncoding>,
    ) -> CosmosMsg {
        let version_metadata = metadata::IcaMetadata::new(
            keys::ICA_VERSION.into(),
            connection_id.clone().into(),
            counterparty_connection_id.into(),
            String::new(),
            tx_encoding.unwrap_or(metadata::TxEncoding::Protobuf),
            "sdk_multi_msg".to_string(),
        );

        let msg_channel_open_init = new_ica_channel_open_init_msg(
            contract_address.clone(),
            format!("wasm.{}", contract_address.into()),
            connection_id,
            counterparty_port_id,
            version_metadata.to_string(),
        );

        CosmosMsg::Stargate {
            type_url: "/ibc.core.channel.v1.MsgChannelOpenInit".into(),
            value: Binary(msg_channel_open_init.encode_to_vec()),
        }
    }

    /// Creates a new MsgChannelOpenInit for an ica channel.
    /// If the counterparty port id is not provided, [`keys::HOST_PORT_ID`] is used.
    fn new_ica_channel_open_init_msg(
        signer: impl Into<String>,
        port_id: impl Into<String>,
        connection_id: impl Into<String>,
        counterparty_port_id: Option<impl Into<String>>,
        version: impl Into<String>,
    ) -> MsgChannelOpenInit {
        let counterparty_port_id = if let Some(port_id) = counterparty_port_id {
            port_id.into()
        } else {
            keys::HOST_PORT_ID.into()
        };

        MsgChannelOpenInit {
            port_id: port_id.into(),
            channel: Some(Channel {
                state: State::Init.into(),
                ordering: Order::Ordered.into(),
                counterparty: Some(Counterparty {
                    port_id: counterparty_port_id,
                    channel_id: String::new(),
                }),
                connection_hops: vec![connection_id.into()],
                version: version.into(),
            }),
            signer: signer.into(),
        }
    }
}
/// Contains the stargate query methods.
pub mod query {
    use super::*;

    use cosmos_sdk_proto::ibc::core::connection::v1::QueryConnectionRequest;
    use cosmwasm_std::{Empty, QuerierWrapper, QueryRequest};

    /// Queries the counterparty connection id using stargate queries.
    pub fn counterparty_connection_id(
        querier: &QuerierWrapper,
        connection_id: impl Into<String>,
    ) -> Result<String, ContractError> {
        let request = QueryConnectionRequest {
            connection_id: connection_id.into(),
        };
        let query: QueryRequest<Empty> = QueryRequest::Stargate {
            path: "/ibc.core.connection.v1.Query/Connection".into(),
            data: Binary(request.encode_to_vec()),
        };

        let response: response::QueryConnectionResponse = querier.query(&query)?;
        Ok(response.connection.counterparty.connection_id)
    }

    /// Contains the types used in query responses.
    mod response {
        /// QueryConnectionResponse is the response type for the Query/Connection RPC
        /// method. Besides the connection end, it includes a proof and the height from
        /// which the proof was retrieved.
        #[derive(Clone, Debug, PartialEq, serde::Deserialize, serde::Serialize)]
        pub struct QueryConnectionResponse {
            /// connection associated with the request identifier
            pub connection: ConnectionEnd,
        }

        /// ConnectionEnd defines a stateful object on a chain connected to another
        /// separate one.
        /// NOTE: there must only be 2 defined ConnectionEnds to establish
        /// a connection between two chains.
        #[derive(Clone, PartialEq, Debug, serde::Deserialize, serde::Serialize)]
        pub struct ConnectionEnd {
            /// client associated with this connection.
            pub client_id: String,
            /// IBC version which can be utilised to determine encodings or protocols for
            /// channels or packets utilising this connection.
            pub versions: Vec<Version>,
            /// current state of the connection end.
            pub state: String,
            /// counterparty chain associated with this connection.
            pub counterparty: Counterparty,
            /// delay period that must pass before a consensus state can be used for
            /// packet-verification NOTE: delay period logic is only implemented by some
            /// clients.
            pub delay_period: String,
        }

        /// Counterparty defines the counterparty chain associated with a connection end.
        #[derive(Clone, PartialEq, Debug, serde::Deserialize, serde::Serialize)]
        pub struct Counterparty {
            /// identifies the client on the counterparty chain associated with a given
            /// connection.
            pub client_id: String,
            /// identifies the connection end on the counterparty chain associated with a
            /// given connection.
            pub connection_id: String,
            /// commitment merkle prefix of the counterparty chain.
            pub prefix: MerklePrefix,
        }

        /// MerklePrefix is merkle path prefixed to the key.
        #[allow(clippy::derive_partial_eq_without_eq)]
        #[derive(Clone, PartialEq, Debug, serde::Deserialize, serde::Serialize)]
        pub struct MerklePrefix {
            /// The constructed key from the Path and the key will be append(Path.KeyPath,
            /// append(Path.KeyPrefix, key...))
            pub key_prefix: String,
        }

        /// Version defines the versioning scheme used to negotiate the IBC verison in
        /// the connection handshake.
        #[derive(Clone, PartialEq, Debug, serde::Deserialize, serde::Serialize)]
        pub struct Version {
            /// unique version identifier
            pub identifier: String,
            /// list of features compatible with the specified identifier
            pub features: Vec<String>,
        }
    }
}
