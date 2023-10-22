//! This file contains helper functions for working with the CwIcaControllerContract.

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{to_binary, Addr, CosmosMsg, StdResult, WasmMsg};

use crate::types::msg;

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
    pub fn call<T: Into<msg::ExecuteMsg>>(&self, msg: T) -> StdResult<CosmosMsg> {
        let msg = to_binary(&msg.into())?;
        Ok(WasmMsg::Execute {
            contract_addr: self.addr().into(),
            msg,
            funds: vec![],
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
    pub fn instantiate<T: Into<msg::InstantiateMsg>>(
        &self,
        msg: T,
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
}
