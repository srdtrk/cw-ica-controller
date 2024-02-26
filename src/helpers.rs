//! This file contains helper functions for working with this contract from
//! external contracts.

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{to_binary, Addr, CosmosMsg, QuerierWrapper, StdResult, WasmMsg};

use crate::types::{msg, state};

pub use cw_ica_controller_derive::ica_callback_execute; // re-export for use in macros

/// `CwIcaControllerContract` is a wrapper around Addr that provides helpers
/// for working with this contract.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct CwIcaControllerContract(pub Addr, pub String);

/// `CwIcaControllerCodeId` is a wrapper around u64 that provides helpers for
/// initializing this contract.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct CwIcaControllerCode(pub u64, pub String);

impl CwIcaControllerContract {
    /// new creates a new [`CwIcaControllerContract`]
    #[must_use]
    pub const fn new(addr: Addr, code_hash: String) -> Self {
        Self(addr, code_hash)
    }

    /// addr returns the address of the contract
    #[must_use]
    pub fn addr(&self) -> Addr {
        self.0.clone()
    }

    /// `code_hash` returns the code hash of the contract
    #[must_use]
    pub fn code_hash(&self) -> String {
        self.1.clone()
    }

    /// call creates a [`WasmMsg::Execute`] message targeting this contract,
    ///
    /// # Errors
    ///
    /// This function returns an error if the given message cannot be serialized
    pub fn call(&self, msg: impl Into<msg::ExecuteMsg>) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Execute {
            contract_addr: self.addr().into(),
            code_hash: self.code_hash(),
            msg,
            funds: vec![],
        }
        .into())
    }

    /// `query_channel` queries the [`state::ChannelState`] of this contract
    ///
    /// # Errors
    ///
    /// This function returns an error if the query fails
    pub fn query_channel(&self, querier: QuerierWrapper) -> StdResult<state::ChannelState> {
        querier.query_wasm_smart(&self.1, &self.0, &msg::QueryMsg::GetChannel {})
    }

    /// `query_state` queries the [`state::ContractState`] of this contract
    ///
    /// # Errors
    ///
    /// This function returns an error if the query fails
    pub fn query_state(&self, querier: QuerierWrapper) -> StdResult<state::ContractState> {
        querier.query_wasm_smart(&self.1, &self.0, &msg::QueryMsg::GetContractState {})
    }

    /// `update_admin` creates a [`WasmMsg::UpdateAdmin`] message targeting this contract
    pub fn update_admin(&self, admin: impl Into<String>) -> CosmosMsg {
        WasmMsg::UpdateAdmin {
            contract_addr: self.addr().into(),
            admin: admin.into(),
        }
        .into()
    }

    /// `clear_admin` creates a [`WasmMsg::ClearAdmin`] message targeting this contract
    #[must_use]
    pub fn clear_admin(&self) -> CosmosMsg {
        WasmMsg::ClearAdmin {
            contract_addr: self.addr().into(),
        }
        .into()
    }
}

impl CwIcaControllerCode {
    /// new creates a new [`CwIcaControllerCode`]
    #[must_use]
    pub const fn new(code_id: u64, code_hash: String) -> Self {
        Self(code_id, code_hash)
    }

    /// `code_id` returns the code id of this code
    #[must_use]
    pub const fn code_id(&self) -> u64 {
        self.0
    }

    /// `code_hash` returns the code hash of this code
    #[must_use]
    pub fn code_hash(&self) -> String {
        self.1.clone()
    }

    /// `instantiate` creates a [`WasmMsg::Instantiate`] message targeting this code
    ///
    /// # Errors
    ///
    /// This function returns an error if the given message cannot be serialized
    pub fn instantiate(
        &self,
        msg: impl Into<msg::InstantiateMsg>,
        label: impl Into<String>,
        admin: Option<impl Into<String>>,
    ) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Instantiate {
            code_id: self.code_id(),
            code_hash: self.code_hash(),
            msg,
            funds: vec![],
            label: label.into(),
            admin: admin.map(Into::into),
        }
        .into())
    }
}
