//! This file contains helper functions for working with this contract from
//! external contracts.

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{to_binary, Addr, Binary, CosmosMsg, QuerierWrapper, StdResult, WasmMsg};

use crate::types::{msg, state};

/// CwIcaControllerContract is a wrapper around Addr that provides helpers
/// for working with this contract.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct CwIcaControllerContract(pub Addr);

/// CwIcaControllerCodeId is a wrapper around u64 that provides helpers for
/// initializing this contract.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, Eq, JsonSchema)]
pub struct CwIcaControllerCode(pub u64);

impl CwIcaControllerContract {
    /// new creates a new [`CwIcaControllerContract`]
    pub fn new(addr: Addr) -> Self {
        Self(addr)
    }

    /// addr returns the address of the contract
    pub fn addr(&self) -> Addr {
        self.0.clone()
    }

    /// call creates a [`WasmMsg::Execute`] message targeting this contract,
    pub fn call(&self, msg: impl Into<msg::ExecuteMsg>) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Execute {
            contract_addr: self.addr().into(),
            msg,
            funds: vec![],
        }
        .into())
    }

    /// query_channel queries the [`state::ChannelState`] of this contract
    pub fn query_channel(&self, querier: QuerierWrapper) -> StdResult<state::ChannelState> {
        querier.query_wasm_smart(self.addr(), &msg::QueryMsg::GetChannel {})
    }

    /// query_state queries the [`state::ContractState`] of this contract
    pub fn query_state(&self, querier: QuerierWrapper) -> StdResult<state::ContractState> {
        querier.query_wasm_smart(self.addr(), &msg::QueryMsg::GetContractState {})
    }

    /// query_callback_counter queries the [`state::CallbackCounter`] of this contract
    pub fn query_callback_counter(
        &self,
        querier: QuerierWrapper,
    ) -> StdResult<state::CallbackCounter> {
        querier.query_wasm_smart(self.addr(), &msg::QueryMsg::GetCallbackCounter {})
    }

    /// update_admin creates a [`WasmMsg::UpdateAdmin`] message targeting this contract
    pub fn update_admin(&self, admin: impl Into<String>) -> StdResult<CosmosMsg> {
        Ok(WasmMsg::UpdateAdmin {
            contract_addr: self.addr().into(),
            admin: admin.into(),
        }
        .into())
    }

    /// clear_admin creates a [`WasmMsg::ClearAdmin`] message targeting this contract
    pub fn clear_admin(&self) -> StdResult<CosmosMsg> {
        Ok(WasmMsg::ClearAdmin {
            contract_addr: self.addr().into(),
        }
        .into())
    }

    /// migrate creates a [`WasmMsg::Migrate`] message targeting this contract
    pub fn migrate(
        &self,
        msg: impl Into<msg::MigrateMsg>,
        new_code_id: u64,
    ) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Migrate {
            contract_addr: self.addr().into(),
            new_code_id,
            msg,
        }
        .into())
    }
}

impl CwIcaControllerCode {
    /// new creates a new [`CwIcaControllerCode`]
    pub fn new(code_id: u64) -> Self {
        Self(code_id)
    }

    /// code_id returns the code id of this code
    pub fn code_id(&self) -> u64 {
        self.0
    }

    /// instantiate creates a [`WasmMsg::Instantiate`] message targeting this code
    pub fn instantiate(
        &self,
        msg: impl Into<msg::InstantiateMsg>,
        label: impl Into<String>,
        admin: Option<impl Into<String>>,
    ) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Instantiate {
            code_id: self.code_id(),
            msg,
            funds: vec![],
            label: label.into(),
            admin: admin.map(|s| s.into()),
        }
        .into())
    }

    /// instantiate2 creates a [`WasmMsg::Instantiate2`] message targeting this code
    pub fn instantiate2(
        &self,
        msg: impl Into<msg::InstantiateMsg>,
        label: impl Into<String>,
        admin: Option<impl Into<String>>,
        salt: impl Into<Binary>,
    ) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Instantiate2 {
            code_id: self.code_id(),
            msg,
            funds: vec![],
            label: label.into(),
            admin: admin.map(|s| s.into()),
            salt: salt.into(),
        }
        .into())
    }
}
