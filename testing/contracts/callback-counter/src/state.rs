use cw_storage_plus::Item;

pub use contract::CallbackCounter;

/// The item used to store the successful and erroneous callbacks in store.
pub const CALLBACK_COUNTER: Item<CallbackCounter> = Item::new("callback_counter");

mod contract {
    use cosmwasm_schema::cw_serde;

    /// CallbackCounter tracks the number of callbacks in store.
    #[cw_serde]
    #[derive(Default)]
    pub struct CallbackCounter {
        /// The number of successful callbacks.
        pub success: u32,
        /// The number of erroneous callbacks.
        pub error: u32,
        /// The number of timeout callbacks.
        /// The channel is closed after a timeout due to the semantics of ordered channels.
        pub timeout: u32,
    }

    impl CallbackCounter {
        /// Increments the success counter
        pub fn success(&mut self) {
            self.success += 1;
        }

        /// Increments the error counter
        pub fn error(&mut self) {
            self.error += 1;
        }

        /// Increments the timeout counter
        pub fn timeout(&mut self) {
            self.timeout += 1;
        }
    }
}
