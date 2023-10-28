package types

type OwnerContract struct {
	Contract
}

func NewOwnerContract(contract Contract) *OwnerContract {
	return &OwnerContract{
		Contract: contract,
	}
}
