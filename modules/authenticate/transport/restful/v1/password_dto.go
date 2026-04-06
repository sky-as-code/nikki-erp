package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
)

type CreateOtpPasswordRequest = it.CreateOtpPasswordCommand

type CreatePasswordOtpResponse struct {
	CreatedAt string `json:"created_at"`
	ExpiredAt string `json:"expired_at"`
	OtpUrl    string `json:"otp_url"`
}

func NewCreateOtpPasswordResponse(data *it.CreatePasswordOtpResultData) CreatePasswordOtpResponse {
	response := CreatePasswordOtpResponse{}
	if data == nil {
		return response
	}

	response.CreatedAt = model.ModelDateTime(data.CreatedAt).String()
	response.ExpiredAt = model.ModelDateTime(data.ExpiredAt).String()
	response.OtpUrl = data.OtpUrl
	return response
}

type ConfirmOtpPasswordRequest = it.ConfirmOtpPasswordCommand

type ConfirmOtpPasswordResponse struct {
	ConfirmedAt   string   `json:"confirmed_at"`
	RecoveryCodes []string `json:"recovery_codes"`
}

func NewConfirmOtpPasswordResponse(data *it.ConfirmOtpPasswordResultData) ConfirmOtpPasswordResponse {
	response := ConfirmOtpPasswordResponse{}
	if data == nil {
		return response
	}

	response.ConfirmedAt = model.ModelDateTime(data.ConfirmedAt).String()
	response.RecoveryCodes = data.RecoveryCodes
	return response
}

type CreateTempPasswordRequest = it.CreateTempPasswordCommand

type CreateTempPasswordResponse struct {
	CreatedAt string `json:"created_at"`
	ExpiredAt string `json:"expired_at"`
}

func NewCreateTempPasswordResponse(data *it.CreateTempPasswordResultData) CreateTempPasswordResponse {
	response := CreateTempPasswordResponse{}
	if data == nil {
		return response
	}

	response.CreatedAt = model.ModelDateTime(data.CreatedAt).String()
	response.ExpiredAt = model.ModelDateTime(data.ExpiredAt).String()
	return response
}

type SetPasswordRequest = it.SetPasswordCommand

type SetPasswordResponse struct {
	UpdatedAt string `json:"updated_at"`
}

func NewSetPasswordResponse(data *it.SetPasswordResultData) SetPasswordResponse {
	response := SetPasswordResponse{}
	if data == nil {
		return response
	}

	response.UpdatedAt = model.ModelDateTime(data.UpdatedAt).String()
	return response
}
