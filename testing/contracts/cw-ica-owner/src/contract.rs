#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
// use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{ContractState, STATE};

/*
// version info for migration info
const CONTRACT_NAME: &str = "crates.io:cw-ica-owner";
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

    STATE.save(
        deps.storage,
        &ContractState::new(admin, msg.ica_controller_code_id),
    )?;
    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::CreateIcaContract {
            salt,
            channel_open_init_options,
        } => execute::create_ica_contract(deps, env, info, salt, channel_open_init_options),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_binary(&query::state(deps)?),
        QueryMsg::GetIcaContractState { ica_id } => to_binary(&query::ica_state(deps, ica_id)?),
        QueryMsg::GetIcaCount {} => to_binary(&query::ica_count(deps)?),
    }
}

mod execute {
    use cosmwasm_std::instantiate2_address;
    use cw_ica_controller::{
        helpers::CwIcaControllerCode, types::msg::options::ChannelOpenInitOptions,
    };

    use crate::state::{self, ICA_COUNT, ICA_STATES};

    use super::*;

    pub fn create_ica_contract(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        salt: Option<String>,
        channel_open_init_options: Option<ChannelOpenInitOptions>,
    ) -> Result<Response, ContractError> {
        let state = STATE.load(deps.storage)?;
        if state.admin != info.sender {
            return Err(ContractError::Unauthorized {});
        }

        let ica_code = CwIcaControllerCode::new(state.ica_controller_code_id);

        let instantiate_msg = cw_ica_controller::types::msg::InstantiateMsg {
            admin: Some(env.contract.address.to_string()),
            channel_open_init_options,
        };

        let ica_count = ICA_COUNT.load(deps.storage).unwrap_or(0);

        let salt = salt.unwrap_or(format!(
            // "test",
            "{}",
            env.block.time.seconds()
        ));
        let label = format!("icacontroller-{}-{}", env.contract.address, ica_count);

        let cosmos_msg = ica_code.instantiate2(
            instantiate_msg,
            label,
            Some(env.contract.address.to_string()),
            salt.as_bytes(),
        )?;

        let code_info = deps
            .querier
            .query_wasm_code_info(state.ica_controller_code_id)?;
        let creator_cannonical = deps.api.addr_canonicalize(env.contract.address.as_str())?;

        let contract_addr = deps.api.addr_humanize(&instantiate2_address(
            &code_info.checksum,
            &creator_cannonical,
            salt.as_bytes(),
        )?)?;

        let initial_state = state::IcaContractState::new(contract_addr);

        ICA_STATES.save(deps.storage, ica_count, &initial_state)?;

        ICA_COUNT.save(deps.storage, &(ica_count + 1))?;

        Ok(Response::new().add_message(cosmos_msg))
    }
}

mod query {
    use crate::state::{IcaContractState, ICA_COUNT, ICA_STATES};

    use super::*;

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        STATE.load(deps.storage)
    }

    /// Returns the saved ICA state for the given ICA ID.
    pub fn ica_state(deps: Deps, ica_id: u64) -> StdResult<IcaContractState> {
        ICA_STATES.load(deps.storage, ica_id)
    }

    /// Returns the saved ICA count.
    pub fn ica_count(deps: Deps) -> StdResult<u64> {
        ICA_COUNT.load(deps.storage)
    }
}

#[cfg(test)]
mod tests {}
