package types

import "fmt"

// newOwnerInstantiateMsg creates a new InstantiateMsg.
func newOwnerInstantiateMsg(admin *string, icaControllerCodeId uint64) string {
	if admin == nil {
		return fmt.Sprintf(`{"ica_controller_code_id":%d}`, icaControllerCodeId)
	} else {
		return fmt.Sprintf(`{"admin":"%s","ica_controller_code_id":%d}`, *admin, icaControllerCodeId)
	}
}

// newOwnerCreateIcaContractMsg creates a new CreateIcaContractMsg.
func newOwnerCreateIcaContractMsg(salt *string) string {
	if salt == nil {
		return `{"create_ica_contract":{}}`
	} else {
		return fmt.Sprintf(`{"create_ica_contract":{"salt":"%s"}}`, *salt)
	}
}
