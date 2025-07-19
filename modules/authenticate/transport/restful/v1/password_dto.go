package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
)

type CreateOtpPasswordRequest = it.CreateOtpPasswordCommand

type CreatePasswordOtpResponse struct {
	CreatedAt int64  `json:"createdAt"`
	ExpiredAt int64  `json:"expiredAt"`
	OtpUrl    string `json:"otpUrl"`
}

func NewCreateOtpPasswordResponse(result it.CreateOtpPasswordResult) CreatePasswordOtpResponse {
	response := CreatePasswordOtpResponse{}
	model.MustCopy(result.Data, &response)
	return response
}

type ConfirmOtpPasswordRequest = it.ConfirmOtpPasswordCommand

type ConfirmOtpPasswordResponse struct {
	ConfirmedAt   int64    `json:"confirmedAt"`
	RecoveryCodes []string `json:"recoveryCodes"`
}

func NewConfirmOtpPasswordResponse(result it.ConfirmOtpPasswordResult) ConfirmOtpPasswordResponse {
	response := ConfirmOtpPasswordResponse{}
	model.MustCopy(result.Data, &response)
	return response
}

type CreateTempPasswordRequest = it.CreateTempPasswordCommand

type CreateTempPasswordResponse struct {
	CreatedAt int64 `json:"createdAt"`
	ExpiredAt int64 `json:"expiredAt"`
}

func NewCreateTempPasswordResponse(result it.CreateTempPasswordResult) CreateTempPasswordResponse {
	response := CreateTempPasswordResponse{}
	model.MustCopy(result.Data, &response)
	return response
}

type SetPasswordRequest = it.SetPasswordCommand

type SetPasswordResponse struct {
	UpdatedAt int64 `json:"updatedAt"`
}

func NewSetPasswordResponse(result it.SetPasswordResult) SetPasswordResponse {
	response := SetPasswordResponse{}
	model.MustCopy(result.Data, &response)
	return response
}
