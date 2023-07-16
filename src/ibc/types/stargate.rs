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

use crate::types::ContractError;
use cosmwasm_std::Binary;

/// Contains the stargate query methods.
pub mod query {
    use super::*;

    use cosmos_sdk_proto::{ibc::core::connection::v1::QueryConnectionRequest, traits::Message};
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
