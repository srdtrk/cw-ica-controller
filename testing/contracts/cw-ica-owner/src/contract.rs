#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
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
        ExecuteMsg::SendPredefinedAction { ica_id, to_address } => {
            execute::send_predefined_action(deps, info, ica_id, to_address)
        }
        ExecuteMsg::ReceiveIcaCallback(callback_msg) => {
            execute::ica_callback_handler(deps, info, callback_msg)
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_json_binary(&query::state(deps)?),
        QueryMsg::GetIcaContractState { ica_id } => {
            to_json_binary(&query::ica_state(deps, ica_id)?)
        }
        QueryMsg::GetIcaCount {} => to_json_binary(&query::ica_count(deps)?),
    }
}

mod execute {
    use cosmwasm_std::{instantiate2_address, Addr};
    use cw_ica_controller::helpers::CwIcaControllerContract;
    use cw_ica_controller::ibc::types::packet::IcaPacketData;
    use cw_ica_controller::types::callbacks::IcaControllerCallbackMsg;
    use cw_ica_controller::types::msg::ExecuteMsg as IcaControllerExecuteMsg;
    use cw_ica_controller::types::state::{ChannelState, ChannelStatus};
    use cw_ica_controller::{
        helpers::CwIcaControllerCode, ibc::types::metadata::TxEncoding,
        types::msg::options::ChannelOpenInitOptions,
    };

    use cosmos_sdk_proto::cosmos::{bank::v1beta1::MsgSend, base::v1beta1::Coin};
    use cosmos_sdk_proto::Any;

    use crate::cosmos_msg::ExampleCosmosMessages;
    use crate::state::{self, CONTRACT_ADDR_TO_ICA_ID, ICA_COUNT, ICA_STATES};

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
            owner: Some(env.contract.address.to_string()),
            channel_open_init_options,
            send_callbacks_to: Some(env.contract.address.to_string()),
        };

        let ica_count = ICA_COUNT.load(deps.storage).unwrap_or(0);

        let salt = salt.unwrap_or(env.block.time.seconds().to_string());
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

        let initial_state = state::IcaContractState::new(contract_addr.clone());

        ICA_STATES.save(deps.storage, ica_count, &initial_state)?;

        CONTRACT_ADDR_TO_ICA_ID.save(deps.storage, contract_addr, &ica_count)?;

        ICA_COUNT.save(deps.storage, &(ica_count + 1))?;

        Ok(Response::new().add_message(cosmos_msg))
    }

    /// Sends a predefined action to the ICA host.
    pub fn send_predefined_action(
        deps: DepsMut,
        info: MessageInfo,
        ica_id: u64,
        to_address: String,
    ) -> Result<Response, ContractError> {
        let contract_state = STATE.load(deps.storage)?;
        contract_state.verify_admin(info.sender)?;

        let ica_state = ICA_STATES.load(deps.storage, ica_id)?;

        let ica_info = if let Some(ica_info) = ica_state.ica_state {
            ica_info
        } else {
            return Err(ContractError::IcaInfoNotSet {});
        };

        let cw_ica_contract = CwIcaControllerContract::new(Addr::unchecked(&ica_state.contract_addr));

        let ica_packet = match ica_info.tx_encoding {
            TxEncoding::Protobuf => {
                let predefined_proto_message = MsgSend {
                    from_address: ica_info.ica_addr,
                    to_address,
                    amount: vec![Coin {
                        denom: "stake".to_string(),
                        amount: "100".to_string(),
                    }],
                };
                IcaPacketData::from_proto_anys(
                    vec![Any::from_msg(&predefined_proto_message)?],
                    None,
                )
            }
            TxEncoding::Proto3Json => {
                let predefined_json_message = ExampleCosmosMessages::MsgSend {
                    from_address: ica_info.ica_addr,
                    to_address,
                    amount: cosmwasm_std::coins(100, "stake"),
                }
                .to_string();
                IcaPacketData::from_json_strings(&[predefined_json_message], None)
            }
        };

        let ica_controller_msg = IcaControllerExecuteMsg::SendCustomIcaMessages {
            messages: Binary(ica_packet.data),
            packet_memo: ica_packet.memo,
            timeout_seconds: None,
        };

        let msg = cw_ica_contract.call(ica_controller_msg)?;

        Ok(Response::default().add_message(msg))
    }

    /// Handles ICA controller callback messages.
    pub fn ica_callback_handler(
        deps: DepsMut,
        info: MessageInfo,
        callback_msg: IcaControllerCallbackMsg,
    ) -> Result<Response, ContractError> {
        let ica_id = CONTRACT_ADDR_TO_ICA_ID.load(deps.storage, info.sender)?;
        let mut ica_state = ICA_STATES.load(deps.storage, ica_id)?;

        match callback_msg {
            IcaControllerCallbackMsg::OnChannelOpenAckCallback {
                channel,
                ica_address,
                tx_encoding,
            } => {
                ica_state.ica_state = Some(state::IcaState {
                    ica_id,
                    channel_state: ChannelState {
                        channel,
                        channel_status: ChannelStatus::Open,
                    },
                    ica_addr: ica_address,
                    tx_encoding,
                });

                ICA_STATES.save(deps.storage, ica_id, &ica_state)?;
            }
            _ => (),
        }

        Ok(Response::default())
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
