/* Code generated by github.com/srdtrk/go-codegen, DO NOT EDIT. */
package simplecounter

type InstantiateMsg struct {
	Count int `json:"count"`
}

type ExecuteMsg struct {
	Increment *ExecuteMsg_Increment `json:"increment,omitempty"`
	Reset *ExecuteMsg_Reset `json:"reset,omitempty"`
}

type QueryMsg struct {
	GetCount *QueryMsg_GetCount `json:"get_count,omitempty"`
}

type QueryMsg_GetCount struct{}

type GetCountResponse struct {
	Count int `json:"count"`
}

type ExecuteMsg_Increment struct{}

type ExecuteMsg_Reset struct {
	Count int `json:"count"`
}
