//! This module handles the execution logic of the contract.

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};

use crate::types::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::types::state::{
    CallbackCounter, ChannelState, ContractState, CALLBACK_COUNTER, CHANNEL_STATE, STATE,
};
use crate::types::ContractError;

// version info for migration
const CONTRACT_NAME: &str = "crates.io:cw-ica-controller";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

/// Instantiates the contract.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    cw2::set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let admin = if let Some(admin) = msg.admin {
        deps.api.addr_validate(&admin)?
    } else {
        info.sender
    };

    // Save the admin. Ica address is determined during handshake.
    STATE.save(deps.storage, &ContractState::new(admin))?;
    // Initialize the callback counter.
    CALLBACK_COUNTER.save(deps.storage, &CallbackCounter::default())?;

    Ok(Response::default())
}

/// Handles the execution of the contract.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::SendCustomIcaMessages {
            messages,
            packet_memo,
            timeout_seconds,
        } => execute::send_custom_ica_messages(
            deps,
            env,
            info,
            messages,
            packet_memo,
            timeout_seconds,
        ),
        ExecuteMsg::SendPredefinedAction { to_address } => {
            execute::send_predefined_action(deps, env, info, to_address)
        }
    }
}

/// Handles the query of the contract.
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_binary(&query::state(deps)?),
        QueryMsg::GetChannel {} => to_binary(&query::channel(deps)?),
        QueryMsg::GetCallbackCounter {} => to_binary(&query::callback_counter(deps)?),
    }
}

mod execute {
    use cosmwasm_std::coins;

    use crate::{ibc::types::packet::IcaPacketData, types::cosmos_msg::ExampleCosmosMessages};

    use super::*;

    // Sends custom messages to the ICA host.
    pub fn send_custom_ica_messages(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        messages: Vec<Binary>,
        packet_memo: Option<String>,
        timeout_seconds: Option<u64>,
    ) -> Result<Response, ContractError> {
        let contract_state = STATE.load(deps.storage)?;
        contract_state.verify_admin(info.sender)?;

        let ica_info = contract_state.get_ica_info()?;
        let ica_messages: Result<Vec<String>, _> = messages
            .into_iter()
            .map(|msg| String::from_utf8(msg.0))
            .collect();

        let ica_packet = IcaPacketData::from_json_strings(ica_messages?, packet_memo)?;
        let send_packet_msg = ica_packet.to_ibc_msg(&env, ica_info.channel_id, timeout_seconds)?;

        Ok(Response::default().add_message(send_packet_msg))
    }

    /// Sends a predefined action to the ICA host.
    pub fn send_predefined_action(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        to_address: String,
    ) -> Result<Response, ContractError> {
        let contract_state = STATE.load(deps.storage)?;
        contract_state.verify_admin(info.sender)?;

        let ica_info = contract_state.get_ica_info()?;
        let predefined_message = ExampleCosmosMessages::MsgSend {
            from_address: ica_info.ica_address,
            to_address,
            amount: coins(100, "stake"),
        }
        .to_string();
        let ica_packet = IcaPacketData::from_json_strings(vec![predefined_message], None)?;
        let send_packet_msg = ica_packet.to_ibc_msg(&env, &ica_info.channel_id, None)?;

        Ok(Response::default().add_message(send_packet_msg))
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

    /// Returns the saved callback counter.
    pub fn callback_counter(deps: Deps) -> StdResult<CallbackCounter> {
        CALLBACK_COUNTER.load(deps.storage)
    }
}

#[cfg(test)]
mod tests {
    use crate::ibc::types::{metadata::TxEncoding, packet::IcaPacketData};
    use crate::types::cosmos_msg::ExampleCosmosMessages;

    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{coins, SubMsg};

    #[test]
    fn test_instantiate() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);

        let msg = InstantiateMsg { admin: None };

        // Ensure the contract is instantiated successfully
        let res = instantiate(deps.as_mut(), env, info.clone(), msg).unwrap();
        assert_eq!(0, res.messages.len());

        // Ensure the admin is saved correctly
        let state = STATE.load(&deps.storage).unwrap();
        assert_eq!(state.admin, info.sender);

        // Ensure the callback counter is initialized correctly
        let counter = CALLBACK_COUNTER.load(&deps.storage).unwrap();
        assert_eq!(counter.success, 0);
        assert_eq!(counter.error, 0);
        assert_eq!(counter.timeout, 0);

        // Ensure that the contract name and version are saved correctly
        let contract_version = cw2::get_contract_version(&deps.storage).unwrap();
        assert_eq!(contract_version.contract, CONTRACT_NAME);
        assert_eq!(contract_version.version, CONTRACT_VERSION);
    }

    #[test]
    fn test_execute_send_custom_ica_messages() {
        let mut deps = mock_dependencies();

        let env = mock_env();
        let info = mock_info("creator", &[]);

        // Instantiate the contract
        let _res = instantiate(
            deps.as_mut(),
            env.clone(),
            info.clone(),
            InstantiateMsg { admin: None },
        )
        .unwrap();

        // for this unit test, we have to set ica info manually or else the contract will error
        STATE
            .update(&mut deps.storage, |mut state| -> StdResult<ContractState> {
                state.set_ica_info("ica_address", "channel-0", TxEncoding::Proto3Json);
                Ok(state)
            })
            .unwrap();

        // Ensure the contract admin can send custom messages
        let custom_msg_str = r#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#;
        let base64_msg = base64::encode(custom_msg_str.as_bytes());

        let messages = vec![Binary::from_base64(&base64_msg).unwrap()];
        let msg = ExecuteMsg::SendCustomIcaMessages {
            messages,
            packet_memo: None,
            timeout_seconds: None,
        };
        let res = execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        let expected_packet =
            IcaPacketData::from_json_strings(vec![custom_msg_str.to_string()], None).unwrap();
        let expected_msg = expected_packet.to_ibc_msg(&env, "channel-0", None).unwrap();

        assert_eq!(1, res.messages.len());
        assert_eq!(res.messages[0], SubMsg::new(expected_msg));

        // Ensure a non-admin cannot send custom messages
        let info = mock_info("non-admin", &[]);
        let msg = ExecuteMsg::SendCustomIcaMessages {
            messages: vec![],
            packet_memo: None,
            timeout_seconds: None,
        };

        let res = execute(deps.as_mut(), env, info, msg);
        assert_eq!(res.unwrap_err().to_string(), "unauthorized".to_string());
    }

    #[test]
    fn test_execute_send_predefined_action() {
        let mut deps = mock_dependencies();

        let env = mock_env();
        let info = mock_info("creator", &[]);

        // Instantiate the contract
        let _res = instantiate(
            deps.as_mut(),
            env.clone(),
            info.clone(),
            InstantiateMsg { admin: None },
        )
        .unwrap();

        // for this unit test, we have to set ica info manually or else the contract will error
        STATE
            .update(&mut deps.storage, |mut state| -> StdResult<ContractState> {
                state.set_ica_info("ica_address", "channel-0", TxEncoding::Proto3Json);
                Ok(state)
            })
            .unwrap();

        // Ensure the contract admin can send predefined messages
        let msg = ExecuteMsg::SendPredefinedAction {
            to_address: "to_address".to_string(),
        };
        let res = execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        let expected_msg = ExampleCosmosMessages::MsgSend {
            from_address: "ica_address".to_string(),
            to_address: "to_address".to_string(),
            amount: coins(100, "stake"),
        }
        .to_string();

        let expected_packet = IcaPacketData::from_json_strings(vec![expected_msg], None).unwrap();
        let expected_msg = expected_packet.to_ibc_msg(&env, "channel-0", None).unwrap();

        assert_eq!(1, res.messages.len());
        assert_eq!(res.messages[0], SubMsg::new(expected_msg));

        // Ensure a non-admin cannot send predefined messages
        let info = mock_info("non-admin", &[]);
        let msg = ExecuteMsg::SendPredefinedAction {
            to_address: "to_address".to_string(),
        };

        let res = execute(deps.as_mut(), env, info, msg);
        assert_eq!(res.unwrap_err().to_string(), "unauthorized".to_string());
    }
}
