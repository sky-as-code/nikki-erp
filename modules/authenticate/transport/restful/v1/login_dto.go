package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
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
	ExpiredAt     string   `json:"expired_at"`
	Methods       []string `json:"methods"`
	SubjectName   string   `json:"subject_name"`
}

func NewStartLoginFlowResponse(data *it.CreateLoginAttemptResultData) StartLoginFlowResponse {
	response := StartLoginFlowResponse{}
	if data == nil {
		return response
	}

	if attemptId := data.Attempt.GetId(); attemptId != nil {
		response.AttemptId = string(*attemptId)
	}
	if createdAt := data.Attempt.GetCreatedAt(); createdAt != nil {
		response.CreatedAt = model.ModelDateTime(*createdAt).String()
	}
	if currentMethod := data.Attempt.GetCurrentMethod(); currentMethod != nil {
		response.CurrentMethod = *currentMethod
	}
	if expiredAt := data.Attempt.GetExpiredAt(); expiredAt != nil {
		response.ExpiredAt = model.ModelDateTime(*expiredAt).String()
	}

	response.Methods = data.Attempt.GetMethods()
	response.SubjectName = data.SubjectName
	return response
}
