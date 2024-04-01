package types

import (
	callbackcounter "github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/callback-counter"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/icacontroller"
	"github.com/srdtrk/cw-ica-controller/interchaintest/v2/types/owner"
)

// CwIcaController_Contract represents the `cw-ica-controller` contract.
type CwIcaController_Contract = Contract[
  icacontroller.InstantiateMsg, icacontroller.ExecuteMsg, icacontroller.QueryMsg,
  ]

// CallbackCounter_Contract represents the `callback-counter` contract.
type CallbackCounter_Contract = Contract[
  callbackcounter.InstantiateMsg, callbackcounter.ExecuteMsg, callbackcounter.QueryMsg,
  ]

// CwIcaOwner_Contract represents the `cw-ica-owner` contract.
type CwIcaOwner_Contract = Contract[
  owner.InstantiateMsg, owner.ExecuteMsg, owner.QueryMsg,
  ]
