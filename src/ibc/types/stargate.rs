//! # Stargate
//!
//! This module contains protobuf types and helpers that are needed to communicate with
//! the core modules of the Cosmos SDK using [`cosmwasm_std::CosmosMsg::Stargate`].

/// Contains the stargate channel lifecycle helper methods.
pub mod channel {
    use cosmwasm_std::{CosmosMsg, IbcOrder};

    use cosmos_sdk_proto::ibc::core::channel::v1::{
        Channel, Counterparty, MsgChannelOpenInit, Order, State,
    };
    use cosmos_sdk_proto::traits::Message;

    use super::super::{keys, metadata};

    /// Creates a new [`MsgChannelOpenInit`] for an ica channel with the given contract address.
    /// Also generates the handshake version.
    /// If the counterparty port id is not provided, [`keys::HOST_PORT_ID`] is used.
    /// If the tx encoding is not provided, [`metadata::TxEncoding::Protobuf`] is used.
    pub fn new_ica_channel_open_init_cosmos_msg(
        contract_address: impl Into<String>,
        connection_id: impl Into<String>,
        counterparty_port_id: Option<impl Into<String>>,
        counterparty_connection_id: impl Into<String>,
        tx_encoding: Option<metadata::TxEncoding>,
        ordering: Option<IbcOrder>,
    ) -> CosmosMsg {
        let contract_address = contract_address.into();
        let connection_id = connection_id.into();

        let version_metadata = metadata::IcaMetadata::new(
            keys::ICA_VERSION.into(),
            connection_id.clone(),
            counterparty_connection_id.into(),
            String::new(),
            tx_encoding.unwrap_or(metadata::TxEncoding::Protobuf),
            "sdk_multi_msg".to_string(),
        );

        let msg_channel_open_init = new_msg_channel_open_init(
            contract_address.clone(),
            format!("wasm.{contract_address}"),
            connection_id,
            counterparty_port_id,
            version_metadata.to_string(),
            ordering,
        );

        #[allow(deprecated)]
        CosmosMsg::Stargate {
            type_url: "/ibc.core.channel.v1.MsgChannelOpenInit".into(),
            value: msg_channel_open_init.encode_to_vec().into(),
        }
    }

    /// Creates a new [`MsgChannelOpenInit`] for an ica channel.
    /// If the counterparty port id is not provided, [`keys::HOST_PORT_ID`] is used.
    fn new_msg_channel_open_init(
        signer: impl Into<String>,
        port_id: impl Into<String>,
        connection_id: impl Into<String>,
        counterparty_port_id: Option<impl Into<String>>,
        version: impl Into<String>,
        ordering: Option<IbcOrder>,
    ) -> MsgChannelOpenInit {
        let counterparty_port_id =
            counterparty_port_id.map_or(keys::HOST_PORT_ID.into(), Into::into);

        let ordering = ordering.map_or(Order::Ordered, |ordering| match ordering {
            IbcOrder::Ordered => Order::Ordered,
            IbcOrder::Unordered => Order::Unordered,
        });

        MsgChannelOpenInit {
            port_id: port_id.into(),
            channel: Some(Channel {
                state: State::Init.into(),
                ordering: ordering.into(),
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
