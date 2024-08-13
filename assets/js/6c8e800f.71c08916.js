"use strict";(self.webpackChunkdocs=self.webpackChunkdocs||[]).push([[4145],{7900:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>l,contentTitle:()=>o,default:()=>p,frontMatter:()=>r,metadata:()=>c,toc:()=>h});var s=t(5893),a=t(1151),i=t(7326);const r={title:"Channel Opening Handshake",sidebar_label:"Channel Opening Handshake",sidebar_position:3,slug:"/how-it-works/channel-handshake"},o="Channel Opening Handshake",c={id:"how-it-works/channel-handshake",title:"Channel Opening Handshake",description:"The channel opening handshake is a 4-step process (see ICS-004 to learn more):",source:"@site/versioned_docs/version-v0.4.x/how-it-works/03-channel-handshake.mdx",sourceDirName:"how-it-works",slug:"/how-it-works/channel-handshake",permalink:"/cw-ica-controller/v0.4/how-it-works/channel-handshake",draft:!1,unlisted:!1,editUrl:"https://github.com/srdtrk/cw-ica-controller/tree/main/docs/versioned_docs/version-v0.4.x/how-it-works/03-channel-handshake.mdx",tags:[],version:"v0.4.x",sidebarPosition:3,frontMatter:{title:"Channel Opening Handshake",sidebar_label:"Channel Opening Handshake",sidebar_position:3,slug:"/how-it-works/channel-handshake"},sidebar:"docsSidebar",previous:{title:"Go vs CosmWasm",permalink:"/cw-ica-controller/v0.4/how-it-works/go-vs-cosmwasm"},next:{title:"Packet Data",permalink:"/cw-ica-controller/v0.4/how-it-works/packet-data"}},l={},h=[{value:"Channel Open Init",id:"channel-open-init",level:2},{value:"Authorization",id:"authorization",level:3},{value:"Version Metadata and Ordering",id:"version-metadata-and-ordering",level:3},{value:"Channel Open Ack",id:"channel-open-ack",level:2}];function d(e){const n={a:"a",admonition:"admonition",code:"code",h1:"h1",h2:"h2",h3:"h3",header:"header",li:"li",ol:"ol",p:"p",pre:"pre",strong:"strong",...(0,a.a)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(n.header,{children:(0,s.jsx)(n.h1,{id:"channel-opening-handshake",children:"Channel Opening Handshake"})}),"\n",(0,s.jsx)(i.Z,{type:"advanced"}),"\n",(0,s.jsx)(i.Z,{type:"protocol"}),"\n",(0,s.jsxs)(n.p,{children:["The channel opening handshake is a 4-step process (see ",(0,s.jsx)(n.a,{href:"https://github.com/cosmos/ibc/tree/main/spec/core/ics-004-channel-and-packet-semantics#opening-handshake",children:"ICS-004"})," to learn more):"]}),"\n",(0,s.jsxs)(n.ol,{children:["\n",(0,s.jsxs)(n.li,{children:[(0,s.jsx)(n.strong,{children:"Channel Open Init"})," (source chain)"]}),"\n",(0,s.jsxs)(n.li,{children:[(0,s.jsx)(n.strong,{children:"Channel Open Try"})," (destination chain)"]}),"\n",(0,s.jsxs)(n.li,{children:[(0,s.jsx)(n.strong,{children:"Channel Open Ack"})," (source chain)"]}),"\n",(0,s.jsxs)(n.li,{children:[(0,s.jsx)(n.strong,{children:"Channel Open Confirm"})," (destination chain)"]}),"\n"]}),"\n",(0,s.jsx)(n.p,{children:"Naturally, this contract only implements the first and third steps of the channel opening handshake, as the second and\nfourth steps are handled by the counterparty ICA host module."}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/v0.4.2/src/ibc/handshake.rs#L15-L46\n"})}),"\n",(0,s.jsx)(n.h2,{id:"channel-open-init",children:"Channel Open Init"}),"\n",(0,s.jsx)(n.h3,{id:"authorization",children:"Authorization"}),"\n",(0,s.jsx)(n.p,{children:"A channel open init message can be sent to any IBC module by any user, and it is up to the module to decide\nwhether to accept the request or not."}),"\n",(0,s.jsxs)(n.p,{children:["In the case of ",(0,s.jsx)(n.code,{children:"cw-ica-controller"}),", only the contract itself can send a channel open init message. Since the sender of\n",(0,s.jsx)(n.code,{children:"MsgChannelOpenInit"})," is not passed to the contract (or any other IBC module), we enforce this by having a ",(0,s.jsx)(n.a,{href:"https://github.com/srdtrk/cw-ica-controller/blob/v0.4.2/src/types/state.rs#L41",children:"state\nvariable"})," that keeps track of whether\nor not to accept channel open init messages. This variable is only set to true by the contract itself right before\nit is about to send a channel open init message in ",(0,s.jsx)(n.a,{href:"/cw-ica-controller/v0.4/contract-api/instantiate-msg",children:(0,s.jsx)(n.code,{children:"InstantiateMsg"})})," or\n",(0,s.jsx)(n.a,{href:"/cw-ica-controller/v0.4/contract-api/execute-msg#createchannel",children:(0,s.jsx)(n.code,{children:"ExecuteMsg::CreateChannel"})}),"."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/v0.4.2/src/contract.rs#L130-L133\n"})}),"\n",(0,s.jsx)(n.h3,{id:"version-metadata-and-ordering",children:"Version Metadata and Ordering"}),"\n",(0,s.jsxs)(n.p,{children:["Whenever a new channel is created, the submitter of ",(0,s.jsx)(n.code,{children:"MsgChannelOpenInit"})," must propose a version string and ordering."]}),"\n",(0,s.jsx)(n.admonition,{type:"info",children:(0,s.jsxs)(n.p,{children:["Interchain Accounts currently only supports ordered channels. This means that a timed-out packet will close the\nchannel. In the upcoming ",(0,s.jsx)(n.code,{children:"ibc-go"})," v8.1 release, unordered channels will be supported, which will allow for the\nICA channels to remain open even after a packet times out."]})}),"\n",(0,s.jsx)(n.p,{children:"In IBC, the version string is used to determine whether or not the two modules on either side of the channel are\ncompatible. The two modules are compatible if and only if they both support and agree on the same version string.\nMoreover, the version string may carry arbitrary metadata encoded in JSON format. This metadata can be used to\ncarry key information about the channel, such as the encoding format, the application version, etc."}),"\n",(0,s.jsxs)(n.p,{children:["The format of the version string for ICS-27 is specified ",(0,s.jsx)(n.a,{href:"https://github.com/cosmos/ibc/tree/main/spec/app/ics-027-interchain-accounts#metadata-negotiation-summary",children:"here"}),".\nThe following rust code shows the version metadata struct used to encode the version string:"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/v0.4.2/src/ibc/types/metadata.rs#L19-L50\n"})}),"\n",(0,s.jsxs)(n.p,{children:["Since it is the contract itself that submits the ",(0,s.jsx)(n.code,{children:"MsgChannelOpenInit"}),", the contract constructs the version string\nbased on the ",(0,s.jsx)(n.code,{children:"channel_open_init_options"})," that are passed to it in ",(0,s.jsx)(n.a,{href:"/cw-ica-controller/v0.4/contract-api/instantiate-msg",children:(0,s.jsx)(n.code,{children:"InstantiateMsg"})})," or ",(0,s.jsx)(n.a,{href:"/cw-ica-controller/v0.4/contract-api/execute-msg#createchannel",children:(0,s.jsx)(n.code,{children:"ExecuteMsg::CreateChannel"})}),"."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/main/src/ibc/types/stargate.rs#L25-L57\n"})}),"\n",(0,s.jsxs)(n.p,{children:["The actual entry point for the ",(0,s.jsx)(n.code,{children:"MsgChannelOpenInit"})," only does validation checks on channel parameters. For example,\nit checks that the channel is not already open."]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/v0.4.2/src/ibc/handshake.rs#L71-L126\n"})}),"\n",(0,s.jsx)(n.h2,{id:"channel-open-ack",children:"Channel Open Ack"}),"\n",(0,s.jsxs)(n.p,{children:["Unlike the ",(0,s.jsx)(n.code,{children:"MsgChannelOpenInit"}),", the ",(0,s.jsx)(n.code,{children:"MsgChannelOpenAck"})," is submitted by a relayer, and we do not need to worry about\nauthorization. This step comes after the ",(0,s.jsx)(n.code,{children:"MsgChannelOpenTry"})," is submitted by the counterparty ICA host module. In the\n",(0,s.jsx)(n.code,{children:"Try"})," step, the counterparty ICA host module may propose a different version string. Therefore, the contract must\nvalidate the version string and channel parameters once again in the ",(0,s.jsx)(n.code,{children:"MsgChannelOpenAck"}),"."]}),"\n",(0,s.jsx)(n.admonition,{type:"note",children:(0,s.jsxs)(n.p,{children:["The interchain account address is passed to the contract in this step through the version string. In ",(0,s.jsx)(n.code,{children:"Init"})," step,\n",(0,s.jsx)(n.code,{children:"cw-ica-controller"})," leaves the interchain account address empty, and the counterparty ICA host module fills it in."]})}),"\n",(0,s.jsx)(n.p,{children:"After validating the version string, the contract then stores the channel parameters in its state."}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-rust",metastring:"reference",children:"https://github.com/srdtrk/cw-ica-controller/blob/v0.4.2/src/ibc/handshake.rs#L128-L188\n"})})]})}function p(e={}){const{wrapper:n}={...(0,a.a)(),...e.components};return n?(0,s.jsx)(n,{...e,children:(0,s.jsx)(d,{...e})}):d(e)}},7326:(e,n,t)=>{t.d(n,{Z:()=>m});var s=t(7294),a=t(394),i=t(512),r=t(8388);function o(){for(var e=arguments.length,n=new Array(e),t=0;t<e;t++)n[t]=arguments[t];return(0,r.m6)((0,i.W)(n))}var c=t(5893);const l=a.fC,h=a.xz,d=s.forwardRef(((e,n)=>{let{className:t,align:s="center",sideOffset:i=4,...r}=e;return(0,c.jsx)(a.VY,{ref:n,align:s,sideOffset:i,className:o("z-50 w-64 rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",t),...r})}));d.displayName=a.VY.displayName;const p={concepts:{color:"#54ffe0",label:"Concepts",isBright:!0,description:"Learn about the concepts behind 'cw-ica-controller'"},basics:{color:"#F69900",label:"Basics",isBright:!0,description:"Learn the basics of 'cw-ica-controller'"},"ibc-go":{color:"#ff1717",label:"IBC-Go",description:"This section includes IBC-Go specific content"},cosmjs:{color:"#6836D0",label:"CosmJS",description:"This section includes CosmJS specific content"},cosmwasm:{color:"#05BDFC",label:"CosmWasm",description:"This section includes CosmWasm specific content"},protocol:{color:"#00B067",label:"Protocol",description:"This section includes content about protocol specifcations"},advanced:{color:"#f7f199",label:"Advanced",isBright:!0,description:"The content in this section is for advanced users researching"},developer:{color:"#AABAFF",label:"Developer",isBright:!0,description:"This section includes content for external developers using the 'cw-ica-controller'"},tutorial:{color:"#F46800",label:"Tutorial",description:"This section includes a tutorial"},"guided-coding":{color:"#F24CF4",label:"Guided Coding",description:"This section includes guided coding"}},m=e=>{let{type:n,version:t}=e;const s=p[n]||p["ibc-go"],a=s.description||"";return(0,c.jsxs)(l,{children:[(0,c.jsx)(h,{children:(0,c.jsxs)("span",{style:{backgroundColor:s.color,borderRadius:"2px",color:s.isBright?"black":"white",padding:"0.3rem",marginBottom:"1rem",marginRight:"0.25rem",display:"inline-block"},children:[s.label,t?` ${t}`:""]})}),(0,c.jsx)(d,{children:a})]})}}}]);