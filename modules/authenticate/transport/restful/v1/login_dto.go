package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
)

type AuthenticateRequest = it.AuthenticateCommand
type AuthenticateResponse = it.AuthenticateResultData

type StartLoginFlowRequest = it.StartLoginFlowCommand

type StartLoginFlowResponse struct {
	AttemptId     string   `json:"attemptId"`
	CreatedAt     int64    `json:"createdAt"`
	CurrentMethod string   `json:"currentMethod"`
	ExpiredAt     int64    `json:"expiredAt"`
	Methods       []string `json:"methods"`
	SubjectName   string   `json:"subjectName"`
}

func NewStartLoginFlowResponse(result it.CreateLoginAttemptResult) StartLoginFlowResponse {
	response := StartLoginFlowResponse{
		AttemptId:   *result.Data.Attempt.Id,
		SubjectName: result.Data.SubjectName,
	}
	model.MustCopy(result.Data.Attempt, &response)
	return response
}
