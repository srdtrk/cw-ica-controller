//! This module handles the execution logic of the contract.

use cosmwasm_std::{entry_point, ContractInfo};
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};

use crate::ibc::types::stargate::channel::new_ica_channel_open_init_cosmos_msg;
use crate::types::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
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
    let owner = msg.owner.unwrap_or_else(|| info.sender.to_string());
    let owner = deps.api.addr_validate(&owner)?;

    state::OWNER.save(deps.storage, &owner)?;

    let callback_contract = msg
        .send_callbacks_to
        .map(|cb| -> StdResult<ContractInfo> {
            Ok(ContractInfo {
                address: deps.api.addr_validate(&cb.address)?,
                code_hash: cb.code_hash,
            })
        })
        .transpose()?;

    // Save the admin. Ica address is determined during handshake.
    state::STATE.save(deps.storage, &ContractState::new(callback_contract))?;

    state::CHANNEL_OPEN_INIT_OPTIONS.save(deps.storage, &msg.channel_open_init_options)?;

    state::ALLOW_CHANNEL_OPEN_INIT.save(deps.storage, &true)?;

    let ica_channel_open_init_msg = new_ica_channel_open_init_cosmos_msg(
        env.contract.address.to_string(),
        msg.channel_open_init_options.connection_id,
        msg.channel_open_init_options.counterparty_port_id,
        msg.channel_open_init_options.counterparty_connection_id,
        msg.channel_open_init_options.tx_encoding,
        msg.channel_open_init_options.channel_ordering,
    );

    Ok(Response::new().add_message(ica_channel_open_init_msg))
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
        ExecuteMsg::CloseChannel {} => execute::close_channel(deps, info),
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
        ExecuteMsg::UpdateCallbackAddress { callback_contract } => {
            execute::update_callback_address(deps, info, callback_contract)
        }
        ExecuteMsg::SendCosmosMsgs {
            messages,
            packet_memo,
            timeout_seconds,
        } => execute::send_cosmos_msgs(deps, env, info, messages, packet_memo, timeout_seconds),
        ExecuteMsg::UpdateOwnership { owner } => execute::update_ownership(deps, info, owner),
    }
}

/// Handles the query of the contract.
#[entry_point]
#[allow(clippy::pedantic)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetContractState {} => to_binary(&query::state(deps)?),
        QueryMsg::GetChannel {} => to_binary(&query::channel(deps)?),
        QueryMsg::Ownership {} => to_binary(&query::ownership(deps.storage)?),
    }
}

mod execute {
    use cosmwasm_std::{ContractInfo, CosmosMsg, IbcMsg, StdResult};

    use crate::{
        ibc::types::packet::IcaPacketData,
        types::msg::{options::ChannelOpenInitOptions, CallbackInfo},
    };

    use super::{
        new_ica_channel_open_init_cosmos_msg, state, Binary, ContractError, DepsMut, Env,
        MessageInfo, Response,
    };

    /// Submits a stargate `MsgChannelOpenInit` to the chain.
    /// Can only be called by the contract owner or a whitelisted address.
    /// Only the contract owner can include the channel open init options.
    #[allow(clippy::needless_pass_by_value)]
    pub fn create_channel(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        options: Option<ChannelOpenInitOptions>,
    ) -> Result<Response, ContractError> {
        state::assert_owner(deps.storage, info.sender)?;

        let options = if let Some(new_options) = options {
            state::CHANNEL_OPEN_INIT_OPTIONS.save(deps.storage, &new_options)?;
            new_options
        } else {
            state::CHANNEL_OPEN_INIT_OPTIONS
                .may_load(deps.storage)?
                .ok_or(ContractError::NoChannelInitOptions)?
        };

        state::ALLOW_CHANNEL_OPEN_INIT.save(deps.storage, &true)?;

        let ica_channel_open_init_msg = new_ica_channel_open_init_cosmos_msg(
            env.contract.address.to_string(),
            options.connection_id,
            options.counterparty_port_id,
            options.counterparty_connection_id,
            options.tx_encoding,
            options.channel_ordering,
        );

        Ok(Response::new().add_message(ica_channel_open_init_msg))
    }

    /// Submits a [`IbcMsg::CloseChannel`].
    #[allow(clippy::needless_pass_by_value)]
    pub fn close_channel(deps: DepsMut, info: MessageInfo) -> Result<Response, ContractError> {
        state::assert_owner(deps.storage, info.sender)?;

        let channel_state = state::CHANNEL_STATE.load(deps.storage)?;
        if !channel_state.is_open() {
            return Err(ContractError::InvalidChannelStatus {
                expected: state::ChannelStatus::Open.to_string(),
                actual: channel_state.channel_status.to_string(),
            });
        }

        state::ALLOW_CHANNEL_CLOSE_INIT.save(deps.storage, &true)?;

        let channel_close_msg = CosmosMsg::Ibc(IbcMsg::CloseChannel {
            channel_id: channel_state.channel.endpoint.channel_id,
        });

        Ok(Response::new().add_message(channel_close_msg))
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
        state::assert_owner(deps.storage, info.sender)?;

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
        state::assert_owner(deps.storage, info.sender)?;

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
        info: MessageInfo,
        owner: String,
    ) -> Result<Response, ContractError> {
        state::assert_owner(deps.storage, info.sender)?;
        let owner = deps.api.addr_validate(&owner)?;

        state::OWNER.save(deps.storage, &owner)?;

        Ok(Response::default())
    }

    /// Updates the callback address.
    #[allow(clippy::needless_pass_by_value)]
    pub fn update_callback_address(
        deps: DepsMut,
        info: MessageInfo,
        callback_contract: Option<CallbackInfo>,
    ) -> Result<Response, ContractError> {
        state::assert_owner(deps.storage, info.sender)?;

        let mut contract_state = state::STATE.load(deps.storage)?;

        contract_state.callback_contract = callback_contract
            .map(|cb| -> StdResult<ContractInfo> {
                Ok(ContractInfo {
                    address: deps.api.addr_validate(&cb.address)?,
                    code_hash: cb.code_hash,
                })
            })
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

    /// Returns the owner of the contract.
    pub fn ownership(storage: &dyn cosmwasm_std::Storage) -> StdResult<String> {
        state::OWNER.load(storage).map(Into::into)
    }
}

#[cfg(test)]
mod tests {
    use crate::ibc::types::{metadata::TxEncoding, packet::IcaPacketData};
    use crate::types::msg::options::ChannelOpenInitOptions;

    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::SubMsg;

    #[test]
    fn test_instantiate() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);

        let channel_open_init_options = ChannelOpenInitOptions {
            connection_id: "connection-0".to_string(),
            counterparty_connection_id: "connection-1".to_string(),
            counterparty_port_id: None,
            tx_encoding: None,
            channel_ordering: None,
        };

        let msg = InstantiateMsg {
            owner: None,
            channel_open_init_options: channel_open_init_options.clone(),
            send_callbacks_to: None,
        };

        let res = instantiate(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        // Ensure that the channel open init options are saved correctly
        assert_eq!(
            state::CHANNEL_OPEN_INIT_OPTIONS
                .load(deps.as_ref().storage)
                .unwrap(),
            channel_open_init_options
        );

        // Ensure the contract is instantiated successfully
        assert_eq!(1, res.messages.len());

        let expected_msg = new_ica_channel_open_init_cosmos_msg(
            env.contract.address.to_string(),
            channel_open_init_options.connection_id,
            channel_open_init_options.counterparty_port_id,
            channel_open_init_options.counterparty_connection_id,
            channel_open_init_options.tx_encoding,
            channel_open_init_options.channel_ordering,
        );
        assert_eq!(res.messages[0], SubMsg::new(expected_msg));

        // Ensure the admin is saved correctly
        let owner = state::OWNER.load(&deps.storage).unwrap();
        assert_eq!(owner, info.sender);
    }

    #[test]
    fn test_execute_send_custom_json_ica_messages() {
        let mut deps = mock_dependencies();

        let env = mock_env();
        let info = mock_info("creator", &[]);

        let channel_open_init_options = ChannelOpenInitOptions {
            connection_id: "connection-0".to_string(),
            counterparty_connection_id: "connection-1".to_string(),
            counterparty_port_id: None,
            tx_encoding: None,
            channel_ordering: None,
        };

        // Instantiate the contract
        let _res = instantiate(
            deps.as_mut(),
            env.clone(),
            info.clone(),
            InstantiateMsg {
                owner: None,
                channel_open_init_options,
                send_callbacks_to: None,
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
            "unauthorized".to_string()
        );
    }
}
