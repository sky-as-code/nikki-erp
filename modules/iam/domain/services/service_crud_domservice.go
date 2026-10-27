package services

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itService "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/service"
)

// NewServiceCrudDomainService is handed the composable default by the iam_service onion, whose
// repository is the user repository viewed through the iam_service schema.
func NewServiceCrudDomainService(
	base composable.CrudDomainService,
	permRepo itPerm.PermissionRepository,
	historyRepo itPerm.PermissionHistoryRepository,
) itService.ServiceCrudDomainService {
	return &ServiceCrudDomainServiceImpl{
		CrudDomainService: base,
		permRepo:          permRepo,
		auditor:           permissionAuditor{historyRepo: historyRepo},
	}
}

// ServiceCrudDomainServiceImpl confines every action to principals of type service (see
// principalTypeScope) and projects only the iam_service fields.
type ServiceCrudDomainServiceImpl struct {
	composable.CrudDomainService
	permRepo itPerm.PermissionRepository
	auditor  permissionAuditor
}

func (this *ServiceCrudDomainServiceImpl) Search(
	ctx corectx.Context, params composable.SearchQuery, options ...corecrud.ServiceSearchOptions,
) (*composable.SearchResult, error) {
	restrictFields(this.Schema(), params)
	if err := serviceTypeScope.narrowSearch(params); err != nil {
		return nil, err
	}
	return this.CrudDomainService.Search(ctx, params, options...)
}

func (this *ServiceCrudDomainServiceImpl) GetById(
	ctx corectx.Context, query composable.GetByIdQuery,
) (*composable.GetOneResult, error) {
	ok, err := serviceTypeScope.isOfType(ctx, this.Repository(), query)
	if err != nil || !ok {
		return &composable.GetOneResult{}, err
	}
	if _, asked := query[fieldsParam]; !asked {
		// The default "whole record" is expanded by the shared repository against iam_user.
		query[fieldsParam] = readableFieldNames(this.Schema())
	}
	restrictFields(this.Schema(), query)
	return this.CrudDomainService.GetById(ctx, query)
}

func (this *ServiceCrudDomainServiceImpl) GetOne(
	ctx corectx.Context, query composable.GetOneQuery,
) (*composable.GetOneResult, error) {
	restrictFields(this.Schema(), query)
	result, err := this.CrudDomainService.GetOne(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	ok, err := serviceTypeScope.isOfType(ctx, this.Repository(), result.Data.Item)
	if err != nil || !ok {
		return &composable.GetOneResult{}, err
	}
	return result, nil
}

// Create writes the type after validation: iam_service declares no `type` field, so validation
// would drop it, but the user repository underneath writes the column.
func (this *ServiceCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.AfterValidationSuccess = setServiceType(opts.AfterValidationSuccess)

	var result *composable.CreateResult
	err := composable.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		result, err = this.CrudDomainService.Create(tranxCtx, cmd, opts)
		if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		// A service holds every system role from its creation, like a user does.
		return this.permRepo.RebuildUserPermission(tranxCtx, composable.ParamId(result.Data))
	})
	if err != nil {
		return nil, errors.Wrap(err, "create service")
	}
	return result, nil
}

func setServiceType(next composable.AfterValidationFn) composable.AfterValidationFn {
	return func(ctx corectx.Context, entity *composable.DynamicEntity) (*composable.DynamicEntity, error) {
		entity.GetFieldData()[domain.UserFieldType] = string(domain.UserTypeService)
		if next == nil {
			return entity, nil
		}
		return next(ctx, entity)
	}
}

func (this *ServiceCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	if refused, err := serviceTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.Update(ctx, cmd, options...)
}

func (this *ServiceCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	if refused, err := serviceTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.Delete(ctx, cmd, options...)
}

func (this *ServiceCrudDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	if refused, err := serviceTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return this.CrudDomainService.SetArchived(ctx, cmd)
}

// ManageRoles writes the same iam_role_user_assignments rows a user's assignment does, then
// rebuilds the service's permission cache in the same transaction.
func (this *ServiceCrudDomainServiceImpl) ManageRoles(
	ctx corectx.Context, cmd itService.ManageServiceRolesCommand,
) (*itService.ManageServiceRolesResult, error) {
	if refused, err := serviceTypeScope.assertOfType(ctx, this.Repository(), cmd); err != nil || refused != nil {
		return refused, err
	}
	return composable.ManageM2m(ctx, this.Repository(), cmd, composable.ManageM2mParam{
		Action:          "manage service role assignments",
		DestSchemaName:  domain.RoleSchemaName,
		JunctionIdField: domain.RoleUserAssignFieldId,
		AfterChange: func(tranxCtx corectx.Context, change composable.M2mChange) error {
			if err := this.permRepo.RebuildUserPermission(tranxCtx, change.SrcId); err != nil {
				return err
			}
			return this.auditor.recordRoleChange(tranxCtx, change.Added, change.Removed,
				[]model.Id{change.SrcId},
				domain.PermissionHistoryReasonRoleAdded, domain.PermissionHistoryReasonRoleRemoved)
		},
	})
}
