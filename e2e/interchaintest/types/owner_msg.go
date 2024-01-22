package types

import (
	"fmt"
)

// newOwnerInstantiateMsg creates a new InstantiateMsg.
func newOwnerInstantiateMsg(admin *string, icaControllerCodeId uint64) string {
	if admin == nil {
		return fmt.Sprintf(`{"ica_controller_code_id":%d}`, icaControllerCodeId)
	} else {
		return fmt.Sprintf(`{"admin":"%s","ica_controller_code_id":%d}`, *admin, icaControllerCodeId)
	}
}
