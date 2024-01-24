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

/// Contains the stargate channel lifecycle helper methods.
pub mod channel {
    use cosmwasm_std::CosmosMsg;

    use cosmos_sdk_proto::traits::Message;
    use cosmos_sdk_proto::ibc::core::channel::v1::{
        Channel, Counterparty, MsgChannelOpenInit, Order, State,
    };

    use super::super::{keys, metadata};

    /// Creates a new [`MsgChannelOpenInit`] for an ica channel with the given contract address.
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
            value: msg_channel_open_init.encode_to_vec().into(),
        }
    }

    /// Creates a new [`MsgChannelOpenInit`] for an ica channel.
    /// If the counterparty port id is not provided, [`keys::HOST_PORT_ID`] is used.
    fn new_ica_channel_open_init_msg(
        signer: impl Into<String>,
        port_id: impl Into<String>,
        connection_id: impl Into<String>,
        counterparty_port_id: Option<impl Into<String>>,
        version: impl Into<String>,
    ) -> MsgChannelOpenInit {
        let counterparty_port_id =
            counterparty_port_id.map_or(keys::HOST_PORT_ID.into(), Into::into);

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
