use cosmwasm_schema::{cw_serde, QueryResponses};

#[cw_serde]
pub struct InstantiateMsg {
    pub admin: Option<String>,
    pub ica_controller_code_id: u64,
}

#[cw_serde]
pub enum ExecuteMsg {
    CreateIcaContract { salt: Option<String> },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {}
