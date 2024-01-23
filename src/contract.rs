//! This module handles the execution logic of the contract.

use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};

use crate::ibc::types::stargate::channel::new_ica_channel_open_init_cosmos_msg;
use crate::types::keys;
use crate::types::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg};
use crate::types::state::{self, ChannelState, ContractState};
use crate::types::ContractError;

/// Instantiates the contract.
#[entry_point]
#[allow(clippy::pedantic)]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    cw2::set_contract_version(deps.storage, keys::CONTRACT_NAME, keys::CONTRACT_VERSION)?;

    let owner = msg.owner.unwrap_or_else(|| info.sender.to_string());
    cw_ownable::initialize_owner(deps.storage, deps.api, Some(&owner))?;

    let callback_address = msg
        .send_callbacks_to
        .map(|addr| deps.api.addr_validate(&addr))
        .transpose()?;

    // Save the admin. Ica address is determined during handshake.
    state::STATE.save(deps.storage, &ContractState::new(callback_address))?;

    if let Some(chan_open_init_whitelist) = msg.channel_open_init_whitelist {
        let chan_open_init_whitelist = chan_open_init_whitelist
            .into_iter()
            .map(|addr| deps.api.addr_validate(&addr))
            .collect::<StdResult<Vec<_>>>()?;
        state::CHANNEL_OPEN_INIT_WHITELIST.save(deps.storage, &chan_open_init_whitelist)?;
    }

    // If channel open init options are provided, open the channel.
    if let Some(channel_open_init_options) = msg.channel_open_init_options {
        state::CHANNEL_OPEN_INIT_OPTIONS.save(deps.storage, &channel_open_init_options)?;

        let ica_channel_open_init_msg = new_ica_channel_open_init_cosmos_msg(
            env.contract.address.to_string(),
            channel_open_init_options.connection_id,
            channel_open_init_options.counterparty_port_id,
            channel_open_init_options.counterparty_connection_id,
            channel_open_init_options.tx_encoding,
        );

        Ok(Response::new().add_message(ica_channel_open_init_msg))
    } else {
        Ok(Response::default())
    }
}

/// Handles the execution of the contract.
#[entry_point]
#[allow(clippy::pedantic)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::CreateChannel {
            channel_open_init_options,
        } => execute::create_channel(deps, env, info, channel_open_init_options),
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
        ExecuteMsg::UpdateCallbackAddress { callback_address } => {
            execute::update_callback_address(deps, info, callback_address)
        }
        ExecuteMsg::SendCosmosMsgs {
            messages,
            packet_memo,
            timeout_seconds,
        } => execute::send_cosmos_msgs(deps, env, info, messages, packet_memo, timeout_seconds),
        ExecuteMsg::UpdateOwnership(action) => execute::update_ownership(deps, env, info, action),
    }
}

/// Handles the query of the contract.
#[entry_point]
#[allow(clippy::pedantic)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_json_binary(&query::state(deps)?),
        QueryMsg::GetChannel {} => to_json_binary(&query::channel(deps)?),
        QueryMsg::Ownership {} => to_json_binary(&cw_ownable::get_ownership(deps.storage)?),
    }
}

/// Migrate contract if version is lower than current version
#[entry_point]
#[allow(clippy::pedantic)]
pub fn migrate(deps: DepsMut, _env: Env, _msg: MigrateMsg) -> Result<Response, ContractError> {
    migrate::validate_semver(deps.as_ref())?;

    cw2::set_contract_version(deps.storage, keys::CONTRACT_NAME, keys::CONTRACT_VERSION)?;
    // If state structure changed in any contract version in the way migration is needed, it
    // should occur here

    Ok(Response::default())
}

mod execute {
    use cosmwasm_std::{CosmosMsg, StdResult};

    use crate::{ibc::types::packet::IcaPacketData, types::msg::options::ChannelOpenInitOptions};

    use super::{
        new_ica_channel_open_init_cosmos_msg, state, Binary, ContractError, DepsMut, Env,
        MessageInfo, Response,
    };

    /// Submits a stargate `MsgChannelOpenInit` to the chain.
    #[allow(clippy::needless_pass_by_value)]
    pub fn create_channel(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        options: Option<ChannelOpenInitOptions>,
    ) -> Result<Response, ContractError> {
        cw_ownable::assert_owner(deps.storage, &info.sender)?;

        state::STATE.update(deps.storage, |mut state| -> StdResult<_> {
            state.enable_channel_open_init();
            Ok(state)
        })?;

        let options = if let Some(new_options) = options {
            state::CHANNEL_OPEN_INIT_OPTIONS.save(deps.storage, &new_options)?;
            new_options
        } else {
            state::CHANNEL_OPEN_INIT_OPTIONS
                .may_load(deps.storage)?
                .ok_or(ContractError::NoChannelInitOptions)?
        };

        let ica_channel_open_init_msg = new_ica_channel_open_init_cosmos_msg(
            env.contract.address.to_string(),
            options.connection_id,
            options.counterparty_port_id,
            options.counterparty_connection_id,
            options.tx_encoding,
        );

        Ok(Response::new().add_message(ica_channel_open_init_msg))
    }

    // Sends custom messages to the ICA host.
    #[allow(clippy::needless_pass_by_value)]
    pub fn send_custom_ica_messages(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        messages: Binary,
        packet_memo: Option<String>,
        timeout_seconds: Option<u64>,
    ) -> Result<Response, ContractError> {
        cw_ownable::assert_owner(deps.storage, &info.sender)?;

        let contract_state = state::STATE.load(deps.storage)?;
        let ica_info = contract_state.get_ica_info()?;

        let ica_packet = IcaPacketData::new(messages.to_vec(), packet_memo);
        let send_packet_msg = ica_packet.to_ibc_msg(&env, ica_info.channel_id, timeout_seconds)?;

        Ok(Response::default().add_message(send_packet_msg))
    }

    /// Sends an array of [`CosmosMsg`] to the ICA host.
    #[allow(clippy::needless_pass_by_value)]
    pub fn send_cosmos_msgs(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        messages: Vec<CosmosMsg>,
        packet_memo: Option<String>,
        timeout_seconds: Option<u64>,
    ) -> Result<Response, ContractError> {
        cw_ownable::assert_owner(deps.storage, &info.sender)?;

        let contract_state = state::STATE.load(deps.storage)?;
        let ica_info = contract_state.get_ica_info()?;

        let ica_packet = IcaPacketData::from_cosmos_msgs(
            messages,
            &ica_info.encoding,
            packet_memo,
            &ica_info.ica_address,
        )?;
        let send_packet_msg = ica_packet.to_ibc_msg(&env, ica_info.channel_id, timeout_seconds)?;

        Ok(Response::default().add_message(send_packet_msg))
    }

    /// Update the ownership of the contract.
    #[allow(clippy::needless_pass_by_value)]
    pub fn update_ownership(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        action: cw_ownable::Action,
    ) -> Result<Response, ContractError> {
        if action == cw_ownable::Action::RenounceOwnership {
            return Err(ContractError::OwnershipCannotBeRenounced);
        };

        cw_ownable::update_ownership(deps, &env.block, &info.sender, action)?;

        Ok(Response::default())
    }

    /// Updates the callback address.
    #[allow(clippy::needless_pass_by_value)]
    pub fn update_callback_address(
        deps: DepsMut,
        info: MessageInfo,
        callback_address: Option<String>,
    ) -> Result<Response, ContractError> {
        cw_ownable::assert_owner(deps.storage, &info.sender)?;

        let mut contract_state = state::STATE.load(deps.storage)?;

        contract_state.callback_address = callback_address
            .map(|addr| deps.api.addr_validate(&addr))
            .transpose()?;

        state::STATE.save(deps.storage, &contract_state)?;

        Ok(Response::default())
    }
}

mod query {
    use super::{state, ChannelState, ContractState, Deps, StdResult};

    /// Returns the saved contract state.
    pub fn state(deps: Deps) -> StdResult<ContractState> {
        state::STATE.load(deps.storage)
    }

    /// Returns the saved channel state if it exists.
    pub fn channel(deps: Deps) -> StdResult<ChannelState> {
        state::CHANNEL_STATE.load(deps.storage)
    }
}

mod migrate {
    use super::{keys, ContractError, Deps};

    pub fn validate_semver(deps: Deps) -> Result<(), ContractError> {
        let prev_cw2_version = cw2::get_contract_version(deps.storage)?;
        if prev_cw2_version.contract != keys::CONTRACT_NAME {
            return Err(ContractError::InvalidMigrationVersion {
                expected: keys::CONTRACT_NAME.to_string(),
                actual: prev_cw2_version.contract,
            });
        }

        let version: semver::Version = keys::CONTRACT_VERSION.parse()?;
        let prev_version: semver::Version = prev_cw2_version.version.parse()?;
        if prev_version >= version {
            return Err(ContractError::InvalidMigrationVersion {
                expected: format!("> {prev_version}"),
                actual: keys::CONTRACT_VERSION.to_string(),
            });
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::ibc::types::{metadata::TxEncoding, packet::IcaPacketData};

    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{Api, SubMsg};

    #[test]
    fn test_instantiate() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);

        let msg = InstantiateMsg {
            owner: None,
            channel_open_init_options: None,
            send_callbacks_to: None,
            channel_open_init_whitelist: None,
        };

        // Ensure the contract is instantiated successfully
        let res = instantiate(deps.as_mut(), env, info.clone(), msg).unwrap();
        assert_eq!(0, res.messages.len());

        // Ensure the admin is saved correctly
        let owner = cw_ownable::get_ownership(&deps.storage)
            .unwrap()
            .owner
            .unwrap();
        assert_eq!(owner, info.sender);

        // Ensure that the contract name and version are saved correctly
        let contract_version = cw2::get_contract_version(&deps.storage).unwrap();
        assert_eq!(contract_version.contract, keys::CONTRACT_NAME);
        assert_eq!(contract_version.version, keys::CONTRACT_VERSION);
    }

    #[test]
    fn test_execute_send_custom_json_ica_messages() {
        let mut deps = mock_dependencies();

        let env = mock_env();
        let info = mock_info("creator", &[]);

        // Instantiate the contract
        let _res = instantiate(
            deps.as_mut(),
            env.clone(),
            info.clone(),
            InstantiateMsg {
                owner: None,
                channel_open_init_options: None,
                send_callbacks_to: None,
                channel_open_init_whitelist: None,
            },
        )
        .unwrap();

        // for this unit test, we have to set ica info manually or else the contract will error
        state::STATE
            .update(&mut deps.storage, |mut state| -> StdResult<ContractState> {
                state.set_ica_info("ica_address", "channel-0", TxEncoding::Proto3Json);
                Ok(state)
            })
            .unwrap();

        // Ensure the contract admin can send custom messages
        let custom_msg_str = r#"{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "to_address": "cosmos15ulrf36d4wdtrtqzkgaan9ylwuhs7k7qz753uk", "amount": [{"denom": "stake", "amount": "5000"}]}"#;
        let messages_str = format!(r#"{{"messages": [{custom_msg_str}]}}"#);
        let base64_json_messages = base64::encode(messages_str.as_bytes());
        let messages = Binary::from_base64(&base64_json_messages).unwrap();

        let msg = ExecuteMsg::SendCustomIcaMessages {
            messages,
            packet_memo: None,
            timeout_seconds: None,
        };
        let res = execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        let expected_packet = IcaPacketData::from_json_strings(&[custom_msg_str.to_string()], None);
        let expected_msg = expected_packet.to_ibc_msg(&env, "channel-0", None).unwrap();

        assert_eq!(1, res.messages.len());
        assert_eq!(res.messages[0], SubMsg::new(expected_msg));

        // Ensure a non-admin cannot send custom messages
        let info = mock_info("non-admin", &[]);
        let msg = ExecuteMsg::SendCustomIcaMessages {
            messages: Binary(vec![]),
            packet_memo: None,
            timeout_seconds: None,
        };

        let res = execute(deps.as_mut(), env, info, msg);
        assert_eq!(
            res.unwrap_err().to_string(),
            "Caller is not the contract's current owner".to_string()
        );
    }

    #[test]
    fn test_update_callback_address() {
        let mut deps = mock_dependencies();

        let env = mock_env();
        let info = mock_info("creator", &[]);

        // Instantiate the contract
        let _res = instantiate(
            deps.as_mut(),
            env.clone(),
            info.clone(),
            InstantiateMsg {
                owner: None,
                channel_open_init_options: None,
                send_callbacks_to: None,
                channel_open_init_whitelist: None,
            },
        )
        .unwrap();

        // Ensure the contract admin can update the callback address
        let new_callback_address = "new_callback_address".to_string();
        let msg = ExecuteMsg::UpdateCallbackAddress {
            callback_address: Some(new_callback_address.clone()),
        };
        let res = execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        assert_eq!(0, res.messages.len());

        let state = state::STATE.load(&deps.storage).unwrap();
        assert_eq!(
            state.callback_address,
            Some(deps.api.addr_validate(&new_callback_address).unwrap())
        );

        // Ensure a non-admin cannot update the callback address
        let info = mock_info("non-admin", &[]);
        let msg = ExecuteMsg::UpdateCallbackAddress {
            callback_address: Some("new_callback_address".to_string()),
        };

        let res = execute(deps.as_mut(), env, info, msg);
        assert_eq!(
            res.unwrap_err().to_string(),
            "Caller is not the contract's current owner".to_string()
        );
    }

    // In this test, we aim to verify that the semver validation is performed correctly.
    // And that the contract version in cw2 is updated correctly.
    #[test]
    fn test_migrate() {
        let mut deps = mock_dependencies();

        let info = mock_info("creator", &[]);

        // Instantiate the contract
        let _res = instantiate(
            deps.as_mut(),
            mock_env(),
            info,
            InstantiateMsg {
                owner: None,
                channel_open_init_options: None,
                send_callbacks_to: None,
                channel_open_init_whitelist: None,
            },
        )
        .unwrap();

        // We need to set the contract version manually to a lower version than the current version
        cw2::set_contract_version(&mut deps.storage, keys::CONTRACT_NAME, "0.0.1").unwrap();

        // Ensure that the contract version is updated correctly
        let contract_version = cw2::get_contract_version(&deps.storage).unwrap();
        assert_eq!(contract_version.contract, keys::CONTRACT_NAME);
        assert_eq!(contract_version.version, "0.0.1");

        // Perform the migration
        let _res = migrate(deps.as_mut(), mock_env(), MigrateMsg {}).unwrap();

        let contract_version = cw2::get_contract_version(&deps.storage).unwrap();
        assert_eq!(contract_version.contract, keys::CONTRACT_NAME);
        assert_eq!(contract_version.version, keys::CONTRACT_VERSION);

        // Ensure that the contract version cannot be downgraded
        cw2::set_contract_version(&mut deps.storage, keys::CONTRACT_NAME, "100.0.0").unwrap();

        let res = migrate(deps.as_mut(), mock_env(), MigrateMsg {});
        assert_eq!(
            res.unwrap_err().to_string(),
            format!(
                "invalid migration version: expected > 100.0.0, got {}",
                keys::CONTRACT_VERSION
            )
        );
    }
}
