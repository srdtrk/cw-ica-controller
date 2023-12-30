package types

type CounterExecuteMsg struct {
	Increment *CounterIncrementMsg `json:"increment,omitempty"`
	Reset     *CounterResetMsg     `json:"reset,omitempty"`
}

type CounterQueryMsg struct {
	GetCount *struct{} `json:"get_count,omitempty"`
}

type GetCountResponse struct {
	Count int64 `json:"count"`
}

type CounterIncrementMsg struct{}

type CounterResetMsg struct {
	Count int64 `json:"count"`
}
