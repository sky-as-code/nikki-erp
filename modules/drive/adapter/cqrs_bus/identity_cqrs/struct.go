package identity_cqrs

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

var existsCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "exists",
}

type UserExistsQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (UserExistsQuery) CqrsRequestType() cqrs.RequestType {
	return existsCommandType
}

func (this UserExistsQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type UserExistsResult = crud.OpResult[bool]
