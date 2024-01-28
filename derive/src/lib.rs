#![doc = include_str!("../Readme.md")]
#![deny(missing_docs)]
#![deny(clippy::nursery, clippy::pedantic, warnings)]

use proc_macro::TokenStream;

use quote::quote;
use syn::{parse_macro_input, AttributeArgs, DataEnum, DeriveInput};

/// Merges the variants of two enums.
/// Adapted from [dao-dao-macros](https://github.com/DA0-DA0/dao-contracts/blob/bc3a44983c1bbad48d12436353a95180489143e8/packages/dao-dao-macros/src/lib.rs)
fn merge_variants(metadata: TokenStream, left: TokenStream, right: TokenStream) -> TokenStream {
    use syn::Data::Enum;

    let args = parse_macro_input!(metadata as AttributeArgs);
    if let Some(first_arg) = args.first() {
        return syn::Error::new_spanned(first_arg, "macro takes no arguments")
            .to_compile_error()
            .into();
    }

    let mut left: DeriveInput = parse_macro_input!(left);
    let right: DeriveInput = parse_macro_input!(right);

    if let (
        Enum(DataEnum { variants, .. }),
        Enum(DataEnum {
            variants: to_add, ..
        }),
    ) = (&mut left.data, right.data)
    {
        variants.extend(to_add);

        quote! { #left }.into()
    } else {
        syn::Error::new(left.ident.span(), "variants may only be added for enums")
            .to_compile_error()
            .into()
    }
}

/// Adds the necessary fields to an enum such that it implements the
/// interface needed to receive callbacks from `cw-ica-controller`.
///
/// For example:
///
/// ```
/// use cw_ica_controller::helper::ica_callback_execute;
/// use cosmwasm_schema::cw_serde;
///
/// #[ica_callback_execute]
/// #[cw_serde]
/// enum ExecuteMsg {}
/// ```
///
/// Will transform the enum to:
///
/// ```
/// enum ExecuteMsg {
///     ReceiveIcaCallback(IcaControllerCallbackMsg),
/// }
/// ```
///
/// Note that other derive macro invocations must occur after this
/// procedural macro as they may depend on the new fields. For
/// example, the following will fail because the `Clone` derivation
/// occurs before the addition of the field.
///
/// ```compile_fail
/// use cw_ica_controller::helper::ica_callback_execute;
/// use cosmwasm_schema::cw_serde;
///
/// #[derive(Clone)]
/// #[ica_callback_execute]
/// #[allow(dead_code)]
/// #[cw_serde]
/// enum Test {
///     Foo,
///     Bar(u64),
///     Baz { foo: u64 },
/// }
/// ```
#[proc_macro_attribute]
pub fn ica_callback_execute(metadata: TokenStream, input: TokenStream) -> TokenStream {
    merge_variants(
        metadata,
        input,
        quote! {
        enum Right {
            /// The callback message from `cw-ica-controller`.
            /// The handler for this variant should verify that this message comes from an
            /// expected legitimate source.
            ReceiveIcaCallback(::cw_ica_controller::types::callbacks::IcaControllerCallbackMsg),
        }
        }
        .into(),
    )
}
