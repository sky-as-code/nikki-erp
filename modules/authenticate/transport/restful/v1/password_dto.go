package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
)

type SetPasswordRequest = it.SetPasswordCommand

type SetPasswordResponse struct {
	UpdatedAt int64 `json:"updatedAt"`
}

func NewSetPasswordResponse(result it.SetPasswordResult) SetPasswordResponse {
	response := SetPasswordResponse{}
	model.MustCopy(result.Data, &response)
	return response
}
