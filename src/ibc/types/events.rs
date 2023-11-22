//! # Events
//!
//! This module defines the events emitted by the ICA controller contract.
//!
//! The core modules will emit events when certain actions occur whether or not
//! the ICA controller contract emits them. This module only defines the events
//! that add more information to the events emitted by the core modules.
//!
//! Therefore;
//!
//! - We need not emit events during the handshake.
//! - We need not emit events during packet sending.
//! - When we emit events associated with packets, it suffices to add attributes
//!   that uniquely identify the packet, and only add attributes that are relevant
//!   to the ICA controller on top of those attributes.

use cosmwasm_std::{Event, IbcPacket};

/// contains the events emitted during packet acknowledgement.
pub mod packet_ack {
    use cosmwasm_std::Binary;

    use super::{attributes, Event, IbcPacket};

    const EVENT_TYPE: &str = "acknowledge_packet";

    /// returns an event for a successful packet acknowledgement.
    #[must_use]
    pub fn success(packet: &IbcPacket, resp: &Binary) -> Event {
        Event::new(EVENT_TYPE)
            .add_attributes(attributes::from_packet(packet))
            .add_attribute(attributes::ACK_BASE64, resp.to_base64())
    }

    /// returns an event for an unsuccessful packet acknowledgement.
    #[must_use]
    pub fn error(packet: &IbcPacket, err: &str) -> Event {
        Event::new(EVENT_TYPE)
            .add_attributes(attributes::from_packet(packet))
            .add_attribute(attributes::ERROR, err)
    }
}

mod attributes {
    use super::IbcPacket;
    use cosmwasm_std::Attribute;

    pub const ACK_BASE64: &str = "packet_ack_base64";
    pub const SEQUENCE: &str = "packet_sequence";
    pub const SRC_PORT: &str = "packet_src_port";
    pub const SRC_CHANNEL: &str = "packet_src_channel";

    pub const ERROR: &str = "error";

    /// returns the attributes for uniquely identifying a packet.
    pub fn from_packet(packet: &IbcPacket) -> Vec<Attribute> {
        vec![
            Attribute::new(SEQUENCE, packet.sequence.to_string()),
            Attribute::new(SRC_PORT, packet.src.port_id.clone()),
            Attribute::new(SRC_CHANNEL, packet.src.channel_id.clone()),
        ]
    }
}
