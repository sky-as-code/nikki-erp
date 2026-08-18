package v1

import (
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/password"
)

type CreateOtpPasswordRequest = it.CreatePasswordOtpCommand

type CreatePasswordOtpResponse struct {
	CreatedAt string `json:"created_at"`
	ExpiredAt string `json:"expired_at"`
	OtpUrl    string `json:"otp_url"`
}

func NewCreateOtpPasswordResponse(data it.CreatePasswordOtpResultData) CreatePasswordOtpResponse {
	response := CreatePasswordOtpResponse{
		CreatedAt: data.CreatedAt.String(),
		ExpiredAt: data.ExpiredAt.String(),
		OtpUrl:    data.OtpUrl,
	}

	return response
}

type ConfirmOtpPasswordRequest = it.ConfirmPasswordOtpCommand

type ConfirmOtpPasswordResponse struct {
	ConfirmedAt   string   `json:"confirmed_at"`
	RecoveryCodes []string `json:"recovery_codes"`
}

func NewConfirmOtpPasswordResponse(data it.ConfirmPasswordOtpResultData) ConfirmOtpPasswordResponse {
	response := ConfirmOtpPasswordResponse{
		ConfirmedAt:   data.ConfirmedAt.String(),
		RecoveryCodes: data.RecoveryCodes,
	}
	return response
}

type CreateTempPasswordRequest = it.CreatePasswordTempCommand

type CreateTempPasswordResponse struct {
	CreatedAt string `json:"created_at"`
	ExpiredAt string `json:"expired_at"`

	// Present only for callers holding `manage_credentials` on iam_user. The
	// omitempty matters: everyone else must see no password key at all.
	Password *string `json:"password,omitempty"`
}

func NewCreateTempPasswordResponse(data it.CreatePasswordTempResultData) CreateTempPasswordResponse {
	response := CreateTempPasswordResponse{
		CreatedAt: data.CreatedAt.String(),
		ExpiredAt: data.ExpiresAt.String(),
		Password:  data.Password,
	}
	return response
}

type SetPasswordRequest = it.SetPasswordCommand

type SetPasswordResponse struct {
	UpdatedAt string `json:"updated_at"`
}

func NewSetPasswordResponse(data dyn.MutateResultData) SetPasswordResponse {
	response := SetPasswordResponse{
		UpdatedAt: data.AffectedAt.String(),
	}
	return response
}
