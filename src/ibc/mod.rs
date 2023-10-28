//! # IBC Module
//!
//! The IBC module is responsible for handling the IBC channel handshake and handling IBC packets.

#[cfg(not(feature = "library"))]
pub mod handshake;
#[cfg(not(feature = "library"))]
pub mod relay;
pub mod types;
