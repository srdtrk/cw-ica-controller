//! # IBC Module
//!
//! The IBC module is responsible for handling the IBC channel handshake and handling IBC packets.

#[cfg(feature = "export")]
pub mod handshake;
#[cfg(feature = "export")]
pub mod relay;
pub mod types;
