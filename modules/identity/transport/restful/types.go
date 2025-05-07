package restful

type RestResponse[TData any] struct {
	Errors []string `json:"errors,omitempty"`
	Data   TData    `json:"data,omitempty"`
}
