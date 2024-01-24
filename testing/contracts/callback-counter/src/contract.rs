#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
// use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{CallbackCounter, CALLBACK_COUNTER};

/*
// version info for migration info
const CONTRACT_NAME: &str = "crates.io:cw-ica-owner";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");
*/

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    CALLBACK_COUNTER.save(deps.storage, &CallbackCounter::default())?;

    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::ReceiveIcaCallback(callback_msg) => {
            execute::ica_callback_handler(deps, info, callback_msg)
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetCallbackCounter {} => to_json_binary(&query::callback_counter(deps)?),
    }
}

mod execute {
    use cw_ica_controller::{
        ibc::types::packet::acknowledgement::Data, types::callbacks::IcaControllerCallbackMsg,
    };

    use crate::state::CALLBACK_COUNTER;

    use super::*;

    /// Handles ICA controller callback messages.
    pub fn ica_callback_handler(
        deps: DepsMut,
        _info: MessageInfo,
        callback_msg: IcaControllerCallbackMsg,
    ) -> Result<Response, ContractError> {
        match callback_msg {
            IcaControllerCallbackMsg::OnChannelOpenAckCallback { .. } => Ok(Response::default()),
            IcaControllerCallbackMsg::OnAcknowledgementPacketCallback {
                ica_acknowledgement,
                ..
            } => match ica_acknowledgement {
                Data::Result(_) => {
                    CALLBACK_COUNTER.update(deps.storage, |mut counter| -> StdResult<_> {
                        counter.success();
                        Ok(counter)
                    })?;
                    Ok(Response::default())
                }
                Data::Error(_) => {
                    CALLBACK_COUNTER.update(deps.storage, |mut counter| -> StdResult<_> {
                        counter.error();
                        Ok(counter)
                    })?;
                    Ok(Response::default())
                }
            },
            IcaControllerCallbackMsg::OnTimeoutPacketCallback { .. } => {
                CALLBACK_COUNTER.update(deps.storage, |mut counter| -> StdResult<_> {
                    counter.timeout();
                    Ok(counter)
                })?;
                Ok(Response::default())
            }
        }
    }
}

mod query {
    use crate::state::{CallbackCounter, CALLBACK_COUNTER};

    use super::*;

    /// Returns the callback counter.
    pub fn callback_counter(deps: Deps) -> StdResult<CallbackCounter> {
        CALLBACK_COUNTER.load(deps.storage)
    }
}

#[cfg(test)]
mod tests {}
