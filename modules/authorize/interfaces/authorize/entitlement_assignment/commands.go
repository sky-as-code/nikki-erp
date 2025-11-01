package entitlement_assignment

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateEntitlementAssignmentCommand)(nil)
	req = (*GetAllEntitlementAssignmentBySubjectQuery)(nil)
	req = (*GetViewsByIdQuery)(nil)
	req = (*GetAllEntitlementAssignmentByEntitlementIdQuery)(nil)
	req = (*DeleteEntitlementAssignmentByIdCommand)(nil)
	req = (*DeleteEntitlementAssignmentByEntitlementIdCommand)(nil)
	req = (*GetByIdQuery)(nil)
	util.Unused(req)
}

// START: CreateEntitlementAssignmentCommand
var createEntitlementAssignmentCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "create",
}

type CreateEntitlementAssignmentCommand struct {
	EntitlementId model.Id                                `json:"entitlementId"`
	SubjectType   domain.EntitlementAssignmentSubjectType `json:"subjectType"`
	SubjectRef    model.Id                                `json:"subjectRef"`
	ActionName    *string                                 `json:"actionName"`
	ResourceName  *string                                 `json:"resourceName"`
	ResolvedExpr  string                                  `json:"resolvedExpr"`
	ScopeRef      *string                                 `json:"scopeRef"`
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

type GetAllEntitlementAssignmentBySubjectResult = crud.OpResult[[]domain.EntitlementAssignment]

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

type GetAllEntitlementAssignmentByEntitlementIdResult = crud.OpResult[[]domain.EntitlementAssignment]

// END: GetAllEntitlementAssignmentByEntitlementIdQuery

// START: DeleteEntitlementAssignmentByIdCommand
var deleteEntitlementAssignmentByIdCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "deleteById",
}

type DeleteEntitlementAssignmentByIdCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (this DeleteEntitlementAssignmentByIdCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (DeleteEntitlementAssignmentByIdCommand) CqrsRequestType() cqrs.RequestType {
	return deleteEntitlementAssignmentByIdCommandType
}

type DeleteEntitlementAssignmentByIdResult = crud.DeletionResult

// END: DeleteEntitlementAssignmentByIdCommand

// START: DeleteEntitlementAssignmentByEntitlementIdCommand
var deleteEntitlementAssignmentByEntitlementIdCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "deleteByEntitlementId",
}

type DeleteEntitlementAssignmentByEntitlementIdCommand struct {
	EntitlementId model.Id `param:"entitlementId" json:"entitlementId"`
}

func (this DeleteEntitlementAssignmentByEntitlementIdCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.EntitlementId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (DeleteEntitlementAssignmentByEntitlementIdCommand) CqrsRequestType() cqrs.RequestType {
	return deleteEntitlementAssignmentByEntitlementIdCommandType
}

type DeleteEntitlementAssignmentByEntitlementIdResult = crud.DeletionResult

// END: DeleteEntitlementAssignmentByEntitlementIdCommand

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

type GetViewsByIdResult = crud.OpResult[[]domain.EntitlementAssignment]

// END: GetViewsByIdQuery

// START: GetByIdQuery
var getByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "getById",
}

type GetByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getByIdQueryType
}

type GetByIdResult = crud.OpResult[*domain.EntitlementAssignment]

// END: GetByIdQuery

// START: GetEntitlementAssignmentByFilterQuery
var getEntitlementAssignmentByFilterQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement_assignment",
	Action:    "getByFilter",
}

type GetEntitlementAssignmentByFilterQuery struct {
	SubjectType   domain.EntitlementAssignmentSubjectType `param:"subjectType" json:"subjectType"`
	SubjectRef    string                                  `param:"subjectRef" json:"subjectRef"`
	EntitlementId model.Id                                `param:"entitlementId" json:"entitlementId"`
	ScopeRef      *string                                 `param:"scopeRef" json:"scopeRef"`
}

func (this GetEntitlementAssignmentByFilterQuery) Validate() fault.ValidationErrors {
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
		model.IdValidateRule(&this.EntitlementId, true),
		validator.Field(&this.ScopeRef,
			validator.When(this.ScopeRef != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_ULID_LENGTH),
			),
		),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetEntitlementAssignmentByFilterQuery) CqrsRequestType() cqrs.RequestType {
	return getEntitlementAssignmentByFilterQueryType
}

type GetEntitlementAssignmentByFilterResult = crud.OpResult[*domain.EntitlementAssignment]

// END: GetEntitlementAssignmentByFilterQuery
