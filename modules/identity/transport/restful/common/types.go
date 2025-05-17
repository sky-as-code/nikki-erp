package common

type RestResponse[TData any] struct {
	Error any   `json:"error,omitempty"`
	Data  TData `json:"data,omitempty"`
}
