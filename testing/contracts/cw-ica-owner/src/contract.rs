#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
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
        ExecuteMsg::CreateIcaContract { salt } => {
            execute::create_ica_contract(deps, env, info, salt)
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(_deps: Deps, _env: Env, _msg: QueryMsg) -> StdResult<Binary> {
    unimplemented!()
}

mod execute {
    use cosmwasm_std::instantiate2_address;
    use cw_ica_controller::helpers::CwIcaControllerCode;

    use crate::state::{self, ICA_COUNT, ICA_STATES};

    use super::*;

    pub fn create_ica_contract(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        salt: Option<String>,
    ) -> Result<Response, ContractError> {
        let state = STATE.load(deps.storage)?;
        if state.admin != info.sender {
            return Err(ContractError::Unauthorized {});
        }

        let ica_code = CwIcaControllerCode::new(state.ica_controller_code_id);

        let instantiate_msg = cw_ica_controller::types::msg::InstantiateMsg {
            admin: Some(env.contract.address.to_string()),
        };

        let ica_count = ICA_COUNT.load(deps.storage).unwrap_or(0);

        let salt = salt.unwrap_or(format!(
            "{}-{}",
            env.contract.address.to_string(),
            env.block.time.nanos()
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
        let code_creator_cannonical = deps.api.addr_canonicalize(&code_info.creator)?;

        let contract_addr = deps.api.addr_humanize(&instantiate2_address(
            &code_info.checksum,
            &code_creator_cannonical,
            salt.as_bytes(),
        )?)?;

        let initial_state = state::IcaContractState::new(contract_addr);

        ICA_STATES.save(deps.storage, ica_count, &initial_state)?;

        ICA_COUNT.save(deps.storage, &(ica_count + 1))?;

        Ok(Response::new().add_message(cosmos_msg))
    }
}

#[cfg(test)]
mod tests {}
