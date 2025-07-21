package entitlement_assignment

import (
	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateEntitlementAssignmentCommand)(nil)
	req = (*GetAllEntitlementAssignmentBySubjectQuery)(nil)
	util.Unused(req)
}

// START: CreateEntitlementAssignmentCommand
var createEntitlementAssignmentCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "create",
}

type CreateEntitlementAssignmentCommand struct {
	SubjectType   *string   `json:"subjectType"`
	SubjectRef    *string   `json:"subjectRef"`
	ActionName    *string   `json:"actionName"`
	ResourceName  *string   `json:"resourceName"`
	ResolvedExpr  *string   `json:"resolvedExpr"`
	EntitlementId *model.Id `json:"entitlementId"`
}

func (CreateEntitlementAssignmentCommand) CqrsRequestType() cqrs.RequestType {
	return createEntitlementAssignmentCommandType
}

type CreateEntitlementAssignmentResult = crud.OpResult[*domain.EntitlementAssignment]

// END: CreateEntitlementAssignmentCommand

// START: GetAllEntitlementAssignmentBySubjectQuery
var getAllEntitlementAssignmentBySubjectQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "getAllBySubject",
}

type GetAllEntitlementAssignmentBySubjectQuery struct {
	SubjectType domain.EntitlementAssignmentSubjectType `param:"subjectType" json:"subjectType"`
	SubjectRef  string                                  `param:"subjectRef" json:"subjectRef"`
}

func (this GetAllEntitlementAssignmentBySubjectQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.SubjectType,
			validator.NotEmpty,
			validator.OneOf(
				domain.EntitlementAssignmentSubjectTypeNikkiUser,
				domain.EntitlementAssignmentSubjectTypeNikkiGroup,
				domain.EntitlementAssignmentSubjectTypeNikkiRole,
				domain.EntitlementAssignmentSubjectTypeCustom,
			),
		),
		model.IdValidateRule(&this.SubjectRef, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetAllEntitlementAssignmentBySubjectQuery) CqrsRequestType() cqrs.RequestType {
	return getAllEntitlementAssignmentBySubjectQueryType
}

type GetAllEntitlementAssignmentBySubjectResult = crud.OpResult[[]*domain.EntitlementAssignment]

// END: GetAllEntitlementAssignmentBySubjectQuery
