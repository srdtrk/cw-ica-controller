//! # Keys
//!
//! Contains key constants definitions for the contract such as version info for migrations.

/// CONTRACT_NAME is the name of the contract recorded with cw2
pub const CONTRACT_NAME: &str = "crates.io:cw-ica-controller";
/// CONTRACT_VERSION is the version of the cargo package.
/// This is also the version of the contract recorded in cw2
pub const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");
