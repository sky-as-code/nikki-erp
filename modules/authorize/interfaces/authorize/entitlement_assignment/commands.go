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
	req = (*GetViewsByIdQuery)(nil)
	req = (*GetAllEntitlementAssignmentByEntitlementIdQuery)(nil)
	req = (*DeleteEntitlementAssignmentByIdQuery)(nil)
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

// START: GetAllEntitlementAssignmentByEntitlementIdQuery

var getAllEntitlementAssignmentByEntitlementIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "getAllByEntitlementId",
}

type GetAllEntitlementAssignmentByEntitlementIdQuery struct {
	EntitlementId model.Id `param:"entitlementId" json:"entitlementId"`
}

func (this GetAllEntitlementAssignmentByEntitlementIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.EntitlementId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetAllEntitlementAssignmentByEntitlementIdQuery) CqrsRequestType() cqrs.RequestType {
	return getAllEntitlementAssignmentByEntitlementIdQueryType
}

type GetAllEntitlementAssignmentByEntitlementIdResult = crud.OpResult[[]*domain.EntitlementAssignment]

// END: GetAllEntitlementAssignmentByEntitlementIdQuery

// START: DeleteEntitlementAssignmentByIdQuery
var deleteEntitlementAssignmentByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "deleteById",
}

type DeleteEntitlementAssignmentByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this DeleteEntitlementAssignmentByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (DeleteEntitlementAssignmentByIdQuery) CqrsRequestType() cqrs.RequestType {
	return deleteEntitlementAssignmentByIdQueryType
}

type DeleteEntitlementAssignmentByIdResult = crud.DeletionResult

// END: DeleteEntitlementAssignmentByIdQuery

// START: GetViewsByIdQuery
var getViewsByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "getViewsById",
}

type GetViewsByIdQuery struct {
	SubjectType string `param:"subjectType" json:"subjectType"`
	SubjectRef  string `param:"subjectRef" json:"subjectRef"`
}

func (this GetViewsByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.SubjectType,
			validator.NotEmpty,
			validator.OneOf(
				domain.EntitlementAssignmentSubjectTypeNikkiUser.String(),
				domain.EntitlementAssignmentSubjectTypeNikkiGroup.String(),
			),
		),
		model.IdValidateRule(&this.SubjectRef, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetViewsByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getViewsByIdQueryType
}

type GetViewsByIdResult = crud.OpResult[[]*domain.EntitlementAssignment]

// END: GetViewsByIdQuery