package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

// NewResourceCrudDomainService is handed the composable default by the iam_resource onion.
func NewResourceCrudDomainService(base composable.CrudDomainService) itRes.ResourceCrudDomainService {
	return &ResourceCrudDomainServiceImpl{CrudDomainService: base}
}

type ResourceCrudDomainServiceImpl struct {
	composable.CrudDomainService
}

func (this *ResourceCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(assertResourceScopeBounds, opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *ResourceCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	opts := composable.UpdateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(assertResourceScopeBounds, opts.ValidateExtra)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// assertResourceScopeBounds refuses a min_scope wider than max_scope. On update a bound the call
// leaves alone keeps its stored value, so either bound may change on its own.
func assertResourceScopeBounds(
	_ corectx.Context, input *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	minScope := effectiveField(input, found, domain.ResourceFieldMinScope)
	maxScope := effectiveField(input, found, domain.ResourceFieldMaxScope)
	if minScope == nil || maxScope == nil {
		return nil
	}
	if domain.AuthzScopeWidth(c.ResourceScope(*minScope)) > domain.AuthzScopeWidth(c.ResourceScope(*maxScope)) {
		vErrs.Append(*ft.NewValidationError(
			domain.ResourceFieldMinScope, ft.ErrorKey("err_min_scope_greater_than_max_scope", "iam"),
			"min_scope must be less than or equal to max_scope"))
	}
	return nil
}

func effectiveField(input *composable.DynamicEntity, found *composable.DynamicEntity, field string) *string {
	if value := input.GetFieldData().GetString(field); value != nil {
		return value
	}
	if found == nil {
		return nil
	}
	return found.GetFieldData().GetString(field)
}

// NewActionCrudDomainService is handed the composable default by the iam_action onion.
func NewActionCrudDomainService(
	base composable.CrudDomainService, resourceRepo itRes.ResourceCrudRepository,
) itAct.ActionCrudDomainService {
	return &ActionCrudDomainServiceImpl{CrudDomainService: base, resourceRepo: resourceRepo}
}

type ActionCrudDomainServiceImpl struct {
	composable.CrudDomainService
	resourceRepo itRes.ResourceCrudRepository
}

func (this *ActionCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(this.assertResourceExists, opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *ActionCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	opts := composable.UpdateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(this.assertResourceExists, opts.ValidateExtra)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// assertResourceExists reports an unknown resource_id as not-found on that field rather than as a
// foreign-key violation.
func (this *ActionCrudDomainServiceImpl) assertResourceExists(
	ctx corectx.Context, input *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	resourceId := input.GetFieldData().GetString(domain.ActionFieldResourceId)
	if resourceId == nil {
		return nil
	}
	exists, err := this.resourceRepo.KeyExist(ctx, dmodel.DynamicFields{basemodel.FieldId: *resourceId})
	if err != nil {
		return err
	}
	if exists.ClientErrors.Count() > 0 || !exists.Data {
		vErrs.Append(*ft.NewNotFoundError(domain.ActionFieldResourceId))
	}
	return nil
}
