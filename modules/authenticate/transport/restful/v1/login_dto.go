package v1

import (
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
)

type AuthenticateRequest = it.AuthenticateCommand
type AuthenticateResponse = it.AuthenticateResultData

type RefreshTokenRequest = it.RefreshTokenCommand
type RefreshTokenResponse = it.RefreshTokenResultData

type StartLoginFlowRequest = it.StartLoginFlowCommand

type StartLoginFlowResponse struct {
	AttemptId     string   `json:"attempt_id"`
	CreatedAt     string   `json:"created_at"`
	CurrentMethod string   `json:"current_method"`
	ExpiresAt     string   `json:"expires_at"`
	Methods       []string `json:"methods"`
	PrincipalName string   `json:"principal_name"`
}

func NewStartLoginFlowResponse(data it.CreateLoginAttemptResultData) StartLoginFlowResponse {
	response := StartLoginFlowResponse{}

	if attemptId := data.Attempt.GetId(); attemptId != nil {
		response.AttemptId = string(*attemptId)
	}
	if createdAt := data.Attempt.GetCreatedAt(); createdAt != nil {
		response.CreatedAt = createdAt.String()
	}
	if currentMethod := data.Attempt.GetCurrentMethod(); currentMethod != nil {
		response.CurrentMethod = *currentMethod
	}
	if expiredAt := data.Attempt.GetExpiresAt(); expiredAt != nil {
		response.ExpiresAt = expiredAt.String()
	}

	response.Methods = data.Attempt.GetMethods()
	response.PrincipalName = data.PrincipalName
	return response
}
