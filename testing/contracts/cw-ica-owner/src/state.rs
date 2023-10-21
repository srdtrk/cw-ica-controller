use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::Item;

pub use contract::ContractState;

/// The item used to store the state of the IBC application.
pub const STATE: Item<ContractState> = Item::new("state");

mod contract {
    use super::*;

    /// ContractState is the state of the IBC application.
    #[cw_serde]
    pub struct ContractState {
        /// The admin of this contract.
        pub admin: Addr,
        /// The code ID of the cw-ica-controller contract.
        pub ica_controller_code_id: u32,
    }

    impl ContractState {
        /// Creates a new ContractState.
        pub fn new(admin: Addr, ica_controller_code_id: u32) -> Self {
            Self {
                admin,
                ica_controller_code_id,
            }
        }
    }
}
