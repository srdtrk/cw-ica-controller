use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

pub use contract::ContractState;
pub use ica::{IcaContractState, IcaState};

/// The item used to store the state of the IBC application.
pub const STATE: Item<ContractState> = Item::new("state");
/// The map used to store the state of the cw-ica-controller contracts.
pub const ICA_STATES: Map<u64, IcaContractState> = Map::new("ica_states");
/// The item used to store the count of the cw-ica-controller contracts.
pub const ICA_COUNT: Item<u64> = Item::new("ica_count");
/// The item used to map contract addresses to ICA IDs.
pub const CONTRACT_ADDR_TO_ICA_ID: Map<Addr, u64> = Map::new("contract_addr_to_ica_id");

mod contract {
    use crate::ContractError;

    use super::*;

    /// ContractState is the state of the IBC application.
    #[cw_serde]
    pub struct ContractState {
        /// The admin of this contract.
        pub admin: Addr,
        /// The code ID of the cw-ica-controller contract.
        pub ica_controller_code_id: u64,
    }

    impl ContractState {
        /// Creates a new ContractState.
        pub fn new(admin: Addr, ica_controller_code_id: u64) -> Self {
            Self {
                admin,
                ica_controller_code_id,
            }
        }

        /// Checks if the address is the admin
        pub fn verify_admin(&self, sender: impl Into<String>) -> Result<(), ContractError> {
            if self.admin == sender.into() {
                Ok(())
            } else {
                Err(ContractError::Unauthorized {})
            }
        }
    }
}

mod ica {
    use cw_ica_controller::{ibc::types::metadata::TxEncoding, types::state::ChannelState};

    use super::*;

    /// IcaContractState is the state of the cw-ica-controller contract.
    #[cw_serde]
    pub struct IcaContractState {
        pub contract_addr: Addr,
        pub ica_state: Option<IcaState>,
    }

    /// IcaState is the state of the ICA.
    #[cw_serde]
    pub struct IcaState {
        pub ica_id: u64,
        pub ica_addr: String,
        pub tx_encoding: TxEncoding,
        pub channel_state: ChannelState,
    }

    impl IcaContractState {
        /// Creates a new [`IcaContractState`].
        pub fn new(contract_addr: Addr) -> Self {
            Self {
                contract_addr,
                ica_state: None,
            }
        }
    }

    impl IcaState {
        /// Creates a new [`IcaState`].
        pub fn new(
            ica_id: u64,
            ica_addr: String,
            tx_encoding: TxEncoding,
            channel_state: ChannelState,
        ) -> Self {
            Self {
                ica_id,
                ica_addr,
                tx_encoding,
                channel_state,
            }
        }
    }
}
