package password

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

var createTempPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "createTempPassword",
}

type CreateTempPasswordCommand struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	SendChannel string             `json:"sendChannel"`
	Username    string             `json:"username"`
}

type CreateTempPasswordResultData struct {
	CreatedAt time.Time `json:"createdAt"`
	ExpiredAt time.Time `json:"expiredAt"`
}
type CreateTempPasswordResult = crud.OpResult[*CreateTempPasswordResultData]

var setPasswordCommandType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "setPassword",
}

type SetPasswordCommand struct {
	SubjectType     domain.SubjectType `json:"subjectType"`
	SubjectRef      model.Id           `json:"subjectRef"`
	CurrentPassword *string            `json:"currentPassword"`
	NewPassword     string             `json:"newPassword"`
}

type SetPasswordResultData struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
type SetPasswordResult = crud.OpResult[*SetPasswordResultData]

var isPasswordMatchedQueryType = cqrs.RequestType{
	Module:    "authenticate",
	Submodule: "password",
	Action:    "isPasswordMatched",
}

type VerifyPasswordQuery struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	Username    string             `json:"username"`
	Password    string             `json:"password"`
}

func (VerifyPasswordQuery) CqrsRequestType() cqrs.RequestType {
	return isPasswordMatchedQueryType
}

type VerifyPasswordResultData struct {
	IsVerified   bool    `json:"isVerified"`
	FailedReason *string `json:"failedReason"`
}
type VerifyPasswordResult = crud.OpResult[*VerifyPasswordResultData]
