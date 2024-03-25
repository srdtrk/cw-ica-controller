"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[6697],{1942:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>r,contentTitle:()=>s,default:()=>h,frontMatter:()=>i,metadata:()=>a,toc:()=>d});var c=t(5893),o=t(1151);const i={title:"InstantiateMsg",sidebar_label:"InstantiateMsg",sidebar_position:1,slug:"/contract-api/instantiate-msg"},s="InstantiateMsg",a={id:"contract-api/instantiate-msg",title:"InstantiateMsg",description:"The InstantiateMsg is the message that is used to instantiate the cw-ica-controller contract.",source:"@site/docs/contract-api/01-instantiate-msg.mdx",sourceDirName:"contract-api",slug:"/contract-api/instantiate-msg",permalink:"/cw-ica-controller/main/contract-api/instantiate-msg",draft:!1,unlisted:!1,editUrl:"https://github.com/srdtrk/cw-ica-controller/tree/main/docs/docs/contract-api/01-instantiate-msg.mdx",tags:[],version:"current",sidebarPosition:1,frontMatter:{title:"InstantiateMsg",sidebar_label:"InstantiateMsg",sidebar_position:1,slug:"/contract-api/instantiate-msg"},sidebar:"docsSidebar",previous:{title:"Overview",permalink:"/cw-ica-controller/main/contract-api/intro"},next:{title:"ExecuteMsg",permalink:"/cw-ica-controller/main/contract-api/execute-msg"}},r={},d=[{value:"Fields",id:"fields",level:2},{value:"<code>owner</code>",id:"owner",level:3},{value:"<code>channel_open_init_options</code>",id:"channel_open_init_options",level:3},{value:"<code>connection_id</code>",id:"connection_id",level:4},{value:"<code>counterparty_connection_id</code>",id:"counterparty_connection_id",level:4},{value:"<code>counterparty_port_id</code>",id:"counterparty_port_id",level:4},{value:"<code>tx_encoding</code>",id:"tx_encoding",level:4},{value:"<code>send_callbacks_to</code>",id:"send_callbacks_to",level:3}];function l(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",h3:"h3",h4:"h4",p:"p",pre:"pre",strong:"strong",...(0,o.a)(),...e.components};return(0,c.jsxs)(c.Fragment,{children:[(0,c.jsx)(n.h1,{id:"instantiatemsg",children:(0,c.jsx)(n.code,{children:"InstantiateMsg"})}),"\n",(0,c.jsxs)(n.p,{children:["The ",(0,c.jsx)(n.code,{children:"InstantiateMsg"})," is the message that is used to instantiate the ",(0,c.jsx)(n.code,{children:"cw-ica-controller"})," contract."]}),"\n",(0,c.jsx)(n.pre,{children:(0,c.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/v0.5.0/src/types/msg.rs#L8-L21\n"})}),"\n",(0,c.jsx)(n.h2,{id:"fields",children:"Fields"}),"\n",(0,c.jsx)(n.h3,{id:"owner",children:(0,c.jsx)(n.code,{children:"owner"})}),"\n",(0,c.jsxs)(n.p,{children:["This contract has an owner who is allowed to call the ",(0,c.jsx)(n.code,{children:"ExecuteMsg"})," methods.\nThe owner management is handled by the amazing ",(0,c.jsx)(n.a,{href:"https://crates.io/crates/cw-ownable",children:"cw-ownable"})," crate.\nIf left empty, the owner is set to the sender of the ",(0,c.jsx)(n.code,{children:"InstantiateMsg"}),"."]}),"\n",(0,c.jsx)(n.h3,{id:"channel_open_init_options",children:(0,c.jsx)(n.code,{children:"channel_open_init_options"})}),"\n",(0,c.jsx)(n.pre,{children:(0,c.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/v0.5.0/src/types/msg.rs#L124-L140\n"})}),"\n",(0,c.jsx)(n.p,{children:"These are the options required for the contract to initiate an ICS-27 channel open handshake.\nThis contract requires there to be an IBC connection between the two chains before it can open a channel."}),"\n",(0,c.jsx)(n.h4,{id:"connection_id",children:(0,c.jsx)(n.code,{children:"connection_id"})}),"\n",(0,c.jsxs)(n.p,{children:["The identifier of the IBC connection end on the deployed (source) chain. (The underlying IBC light client must\nbe live.) If this field is set to a non-existent connection, the execution of the ",(0,c.jsx)(n.code,{children:"InstantiateMsg"})," will fail."]}),"\n",(0,c.jsx)(n.h4,{id:"counterparty_connection_id",children:(0,c.jsx)(n.code,{children:"counterparty_connection_id"})}),"\n",(0,c.jsxs)(n.p,{children:["The identifier of the IBC connection end on the counterparty (destination) chain. (The underlying IBC light\nclient must be live.) If this field is set to a non-existent connection or a different connection's end,\nthen the execution of the ",(0,c.jsx)(n.code,{children:"InstantiateMsg"})," will not fail. This is because the source chain does not know\nabout the counterparty chain's connections. Instead, the channel open handshake will fail to complete."]}),"\n",(0,c.jsxs)(n.p,{children:["If the contract was instantiated with a ",(0,c.jsx)(n.code,{children:"counterparty_connection_id"})," that does not match the connection\nend on the counterparty chain, then the owner must call ",(0,c.jsx)(n.a,{href:"/cw-ica-controller/main/contract-api/execute-msg#createchannel",children:(0,c.jsx)(n.code,{children:"ExecuteMsg::CreateChannel"})})," with the correct parameters to start a new channel open handshake."]}),"\n",(0,c.jsx)(n.h4,{id:"counterparty_port_id",children:(0,c.jsx)(n.code,{children:"counterparty_port_id"})}),"\n",(0,c.jsxs)(n.p,{children:["This is a required parameter for the ICS-27 channel version metadata. I've added it here for consistency.\nCurrently, the only supported value is ",(0,c.jsx)(n.code,{children:"icahost"}),". If left empty, it is set to ",(0,c.jsx)(n.code,{children:"icahost"}),".\n",(0,c.jsx)(n.strong,{children:"So you should ignore this field."})]}),"\n",(0,c.jsx)(n.h4,{id:"tx_encoding",children:(0,c.jsx)(n.code,{children:"tx_encoding"})}),"\n",(0,c.jsxs)(n.p,{children:["The ICS-27 implementation in ",(0,c.jsx)(n.code,{children:"ibc-go"})," supports two transaction encoding formats: ",(0,c.jsx)(n.code,{children:"proto3"})," and ",(0,c.jsx)(n.code,{children:"proto3json"}),".\nThis contract supports both formats, but defaults to ",(0,c.jsx)(n.code,{children:"proto3"}),". This is because most chains don't yet support\n",(0,c.jsx)(n.code,{children:"proto3json"})," encoding. Moreover, this contract does not support some ",(0,c.jsx)(n.code,{children:"CosmosMsg"})," while using ",(0,c.jsx)(n.code,{children:"proto3json"}),".\n",(0,c.jsx)(n.strong,{children:"So, if you are unsure, leave this field empty."})]}),"\n",(0,c.jsx)(n.h3,{id:"send_callbacks_to",children:(0,c.jsx)(n.code,{children:"send_callbacks_to"})}),"\n",(0,c.jsxs)(n.p,{children:["This is the address of the contract that will receive the callbacks from the ",(0,c.jsx)(n.code,{children:"cw-ica-controller"})," contract.\nThis may be the same address as the ",(0,c.jsx)(n.code,{children:"owner"})," or a different address. If left empty, no callbacks will be sent.\nLearn more about callbacks ",(0,c.jsx)(n.a,{href:"/cw-ica-controller/main/contract-api/callbacks",children:"here"}),"."]})]})}function h(e={}){const{wrapper:n}={...(0,o.a)(),...e.components};return n?(0,c.jsx)(n,{...e,children:(0,c.jsx)(l,{...e})}):l(e)}},1151:(e,n,t)=>{t.d(n,{Z:()=>a,a:()=>s});var c=t(7294);const o={},i=c.createContext(o);function s(e){const n=c.useContext(i);return c.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function a(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:s(e.components),c.createElement(i.Provider,{value:n},e.children)}}}]);