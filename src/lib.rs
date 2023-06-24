//! # ICA Controller
//!
//! This module is an implementation of the Interchain Accounts Controller (ICA Controller) in CosmWasm.
//! It is a contract that can be deployed on a chain to control an Interchain Account (ICA) in a host chain.
//! It does this by completing the IBC handshake with the ICA host, and then sending messages to the ICA host.
//!
//! The handshake is handled in the [`ibc`] module. Unlike the golang implementation, this ICA controller:
//!
//! - Can only handle one ICA at a time.
//! - The handshake must be initiated by a relayer.
//! - The version passed in to the `OpenInit` message cannot be the empty string. This is due to the limitations
//!   of the IBCModule interface. See [#3942](https://github.com/cosmos/ibc-go/issues/3942).

#![deny(missing_docs)]

pub mod contract;
pub mod ibc;
pub mod types;
