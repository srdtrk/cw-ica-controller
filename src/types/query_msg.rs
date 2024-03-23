//! This module contains the helpers to convert [`QueryRequest`] to [`cosmos_sdk_proto::Any`]

#[derive(::prost::Message)]
/// `MsgModuleQuerySafe` defines the query request tx added in ibc-go v8.2
pub struct MsgModuleQuerySafe {
    #[prost(string, tag = "1")]
    /// signer is the address of the account that signed the transaction
    pub signer: ::prost::alloc::string::String,
    /// requests is the list of query requests
    #[prost(message, repeated, tag = "2")]
    pub requests: ::prost::alloc::vec::Vec<QueryRequest>,
}

#[derive(::prost::Message)]
/// `QueryRequest` defines the parameters for a particular query request by an interchain account.
pub struct QueryRequest {
    #[prost(string, tag = "1")]
    /// `path` defines the path of the query request as defined by ADR-021.
    /// https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-021-protobuf-query-encoding.md#custom-query-registration-and-routing
    pub path: ::prost::alloc::string::String,
    #[prost(bytes = "vec", tag = "2")]
    /// `data` defines the payload of the query request as defined by ADR-021.
    /// https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-021-protobuf-query-encoding.md#custom-query-registration-and-routing
    pub data: ::prost::alloc::vec::Vec<u8>,
}

#[derive(::prost::Message)]
/// `MsgModuleQuerySafeResponse` defines the response for Msg/ModuleQuerySafe
pub struct MsgModuleQuerySafeResponse {
    /// responses is the list of query responses as bytes
    /// The responses are in the same order as the requests
    #[prost(bytes = "vec", repeated, tag = "1")]
    pub responses: ::prost::alloc::vec::Vec<Vec<u8>>,
}
