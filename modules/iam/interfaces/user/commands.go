package user

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// GetUserQuery reads one user by id or by email. It is the request of the getEnabledUser command
// bus query the request guard sends on every authenticated call: the request type and the JSON
// shape are copied by core, so neither may change.
var getUserByIdQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "user",
	Action:    "getUserById",
}

type GetUserQuery struct {
	Fields []string  `json:"fields" query:"fields"`
	Id     *model.Id `json:"id" param:"id"`
	Email  *string   `json:"email" query:"email"`
}

func (GetUserQuery) CqrsRequestType() cqrs.RequestType {
	return getUserByIdQueryType
}

func (GetUserQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.get_user_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				ExclusiveRequiredFields(domain.UserFieldId, domain.UserFieldEmail).
				Field(dyn.DefineFieldSearchFields()).
				Field(dmodel.DefineField().
					Name(basemodel.FieldId).
					DataType(dmodel.FieldDataTypeUlid())).
				Field(dmodel.DefineField().
					Name(domain.UserFieldEmail).
					DataType(dmodel.FieldDataTypeEmail()))
		},
	)
}

// GetUserDomainResult is the reply of GetUserQuery: one user, typed.
type GetUserDomainResult = dyn.OpResult[domain.User]
