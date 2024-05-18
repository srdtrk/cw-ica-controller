use cw_storage_plus::Item;

pub use contract::CallbackCounter;

/// The item used to store the successful and erroneous callbacks in store.
pub const CALLBACK_COUNTER: Item<CallbackCounter> = Item::new("callback_counter");

mod contract {
    use cosmwasm_schema::cw_serde;
    use cw_ica_controller::types::callbacks::IcaControllerCallbackMsg;

    /// CallbackCounter tracks the number of callbacks in store.
    #[cw_serde]
    #[derive(Default)]
    pub struct CallbackCounter {
        /// The successful callbacks.
        pub success: Vec<IcaControllerCallbackMsg>,
        /// The erroneous callbacks.
        pub error: Vec<IcaControllerCallbackMsg>,
        /// The timeout callbacks.
        /// The channel is closed after a timeout if the channel is ordered due to the semantics of ordered channels.
        pub timeout: Vec<IcaControllerCallbackMsg>,
    }

    impl CallbackCounter {
        /// Pushes to the success counter
        pub fn success(&mut self, msg: IcaControllerCallbackMsg) {
            self.success.push(msg);
        }

        /// Pushes to the error counter
        pub fn error(&mut self, msg: IcaControllerCallbackMsg) {
            self.error.push(msg);
        }

        /// Pushes to the timeout counter
        pub fn timeout(&mut self, msg: IcaControllerCallbackMsg) {
            self.timeout.push(msg);
        }
    }
}
