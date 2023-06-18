#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
// use cw2::set_contract_version;

use crate::types::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::types::state::{ChannelState, ContractState, CHANNEL_STATE, STATE};
use crate::types::ContractError;

/*
// version info for migration info
const CONTRACT_NAME: &str = "crates.io:cw-ica-controller";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");
*/

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    let admin = if let Some(admin) = msg.admin {
        deps.api.addr_validate(&admin)?
    } else {
        info.sender
    };

    // Save the admin. Ica address is determined during handshake.
    STATE.save(deps.storage, &ContractState::new(admin, None))?;

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
        ExecuteMsg::SendCustomIcaMessages { messages } => {
            execute::send_custom_ica_messages(deps, info, messages)
        }
    }
}

mod execute {
    use super::*;

    // send_custom_ica_messages sends custom messages from the ICA controller to the ICA host.
    pub fn send_custom_ica_messages(
        deps: DepsMut,
        info: MessageInfo,
        messages: Vec<String>,
    ) -> Result<Response, ContractError> {
        todo!()
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_binary(&query::state(deps)?),
        QueryMsg::GetChannel {} => to_binary(&query::channel(deps)?),
    }
}

mod query {
    use super::*;

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        STATE.load(deps.storage)
    }

    /// Returns the saved channel state if it exists.
    pub fn channel(deps: Deps) -> StdResult<ChannelState> {
        CHANNEL_STATE.load(deps.storage)
    }
}

#[cfg(test)]
mod tests {}
