package app

import (
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	itAssignment "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/permission_history"
	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewEntitlementServiceImpl(
	actionService itAction.ActionService,
	assignmentService itAssignment.EntitlementAssignmentService,
	cqrsBus cqrs.CqrsBus,
	entitlementRepo it.EntitlementRepository,
	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository,
	resourceService itResource.ResourceService,
) it.EntitlementService {
	return &EntitlementServiceImpl{
		actionService:         actionService,
		assignmentService:     assignmentService,
		cqrsBus:               cqrsBus,
		entitlementRepo:       entitlementRepo,
		permissionHistoryRepo: permissionHistoryRepo,
		resourceService:       resourceService,
	}
}

type EntitlementServiceImpl struct {
	actionService         itAction.ActionService
	assignmentService     itAssignment.EntitlementAssignmentService
	cqrsBus               cqrs.CqrsBus
	entitlementRepo       it.EntitlementRepository
	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository
	resourceService       itResource.ResourceService
}

func (this *EntitlementServiceImpl) CreateEntitlement(ctx crud.Context, cmd it.CreateEntitlementCommand) (*it.CreateEntitlementResult, error) {
	return crud.Create(ctx, crud.CreateParam[*domain.Entitlement, it.CreateEntitlementCommand, it.CreateEntitlementResult]{
		Action:              "create entitlement",
		Command:             cmd,
		AssertBusinessRules: this.assertBusinessRuleCreateEntitlement,
		RepoCreate:          this.entitlementRepo.Create,
		SetDefault:          this.setEntitlementDefaults,
		Sanitize:            this.sanitizeEntitlement,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.CreateEntitlementResult {
			return &it.CreateEntitlementResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Entitlement) *it.CreateEntitlementResult {
			return &it.CreateEntitlementResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})
}

func (this *EntitlementServiceImpl) EntitlementExists(ctx crud.Context, cmd it.EntitlementExistsQuery) (result *it.EntitlementExistsResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "check entitlement exists"); e != nil {
			err = e
		}
	}()

	var existsEntitlement bool

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			existsEntitlement, err = this.entitlementRepo.Exists(ctx, it.FindByIdParam{Id: cmd.Id})
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.EntitlementExistsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.EntitlementExistsResult{
		Data:    existsEntitlement,
		HasData: true,
	}, nil
}

func (this *EntitlementServiceImpl) UpdateEntitlement(ctx crud.Context, cmd it.UpdateEntitlementCommand) (*it.UpdateEntitlementResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Entitlement, it.UpdateEntitlementCommand, it.UpdateEntitlementResult]{
		Action:       "update entitlement",
		Command:      cmd,
		AssertExists: this.assertEntitlementExistsById,
		RepoUpdate:   this.entitlementRepo.Update,
		Sanitize:     this.sanitizeEntitlement,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.UpdateEntitlementResult {
			return &it.UpdateEntitlementResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Entitlement) *it.UpdateEntitlementResult {
			return &it.UpdateEntitlementResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *EntitlementServiceImpl) DeleteEntitlementHard(ctx crud.Context, cmd it.DeleteEntitlementHardByIdCommand) (result *it.DeleteEntitlementHardByIdResult, err error) {
	tx, err := this.entitlementRepo.BeginTransaction(ctx)
	fault.PanicOnErr(err)

	ctx.SetDbTranx(tx)

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if result != nil && result.ClientError != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	result, err = crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Entitlement, it.DeleteEntitlementHardByIdCommand, it.DeleteEntitlementHardByIdResult]{
		Action:              "delete entitlement",
		Command:             cmd,
		AssertExists:        this.assertEntitlementExistsById,
		AssertBusinessRules: this.assertBusinessRuleDeleteEntitlement,
		RepoDelete: func(ctx crud.Context, model *domain.Entitlement) (int, error) {
			return this.entitlementRepo.DeleteHard(ctx, it.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteEntitlementHardByIdResult {
			return &it.DeleteEntitlementHardByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Entitlement, deletedCount int) *it.DeleteEntitlementHardByIdResult {
			return crud.NewSuccessDeletionResult(*model.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *EntitlementServiceImpl) GetEntitlementById(ctx crud.Context, query it.GetEntitlementByIdQuery) (*it.GetEntitlementByIdResult, error) {
	return crud.GetOne(ctx, crud.GetOneParam[*domain.Entitlement, it.GetEntitlementByIdQuery, it.GetEntitlementByIdResult]{
		Action:      "get entitlement by Id",
		Query:       query,
		RepoFindOne: this.getEntitlementByIdFull,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.GetEntitlementByIdResult {
			return &it.GetEntitlementByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Entitlement) *it.GetEntitlementByIdResult {
			return &it.GetEntitlementByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})
}

func (this *EntitlementServiceImpl) GetAllEntitlementByIds(ctx crud.Context, query it.GetAllEntitlementByIdsQuery) (result *it.GetAllEntitlementByIdsResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get all entitlement by ids"); e != nil {
			err = e
		}
	}()

	var dbEntitlements []domain.Entitlement
	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbEntitlements, err = this.entitlementRepo.FindAllByIds(ctx, query)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetAllEntitlementByIdsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if len(dbEntitlements) == 0 {
		vErrs.AppendNotFound("id", "entitlement")

		return &it.GetAllEntitlementByIdsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetAllEntitlementByIdsResult{
		Data:    dbEntitlements,
		HasData: dbEntitlements != nil,
	}, nil
}

func (this *EntitlementServiceImpl) SearchEntitlements(ctx crud.Context, query it.SearchEntitlementsQuery) (*it.SearchEntitlementsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Entitlement, it.SearchEntitlementsQuery, it.SearchEntitlementsResult]{
		Action: "search entitlements",
		Query:  query,
		SetQueryDefaults: func(query *it.SearchEntitlementsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.entitlementRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query it.SearchEntitlementsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Entitlement], error) {
			return this.entitlementRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.SearchEntitlementsResult {
			return &it.SearchEntitlementsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Entitlement]) *it.SearchEntitlementsResult {
			return &it.SearchEntitlementsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *EntitlementServiceImpl) getEntitlementByIdFull(ctx crud.Context, query it.GetEntitlementByIdQuery, vErrs *fault.ValidationErrors) (dbEntitlement *domain.Entitlement, err error) {
	dbEntitlement, err = this.entitlementRepo.FindById(ctx, it.FindByIdParam{Id: query.Id})
	fault.PanicOnErr(err)

	if dbEntitlement == nil {
		vErrs.AppendNotFound("entitlement_id", "entitlement")
	}
	return
}

func (this *EntitlementServiceImpl) assertEntitlementExistsById(ctx crud.Context, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) (dbEntitlement *domain.Entitlement, err error) {
	dbEntitlement, err = this.entitlementRepo.FindById(ctx, it.FindByIdParam{Id: *entitlement.Id})
	fault.PanicOnErr(err)

	if dbEntitlement == nil {
		vErrs.AppendNotFound("entitlement_id", "entitlement")
	}
	return
}

func (this *EntitlementServiceImpl) sanitizeEntitlement(entitlement *domain.Entitlement) {
	if entitlement.Description != nil {
		entitlement.Description = util.ToPtr(defense.SanitizePlainText(*entitlement.Description, true))
	}

	if entitlement.Name != nil {
		entitlement.Name = util.ToPtr(defense.SanitizePlainText(*entitlement.Name, true))
	}
}

func (this *EntitlementServiceImpl) setEntitlementDefaults(entitlement *domain.Entitlement) {
	entitlement.SetDefaults()
}

func (this *EntitlementServiceImpl) assertEntitlementUnique(ctx crud.Context, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	if entitlement.Name == nil {
		return nil
	}

	dbEntitlement, err := this.entitlementRepo.FindByName(
		ctx,
		it.FindByNameParam{
			Name:  *entitlement.Name,
			OrgId: entitlement.OrgId,
		},
	)
	fault.PanicOnErr(err)

	if dbEntitlement != nil {
		vErrs.AppendAlreadyExists("entitlement_name", "entitlement name")
	}
	return nil
}

func (this *EntitlementServiceImpl) assertActionExprValid(resource *domain.Resource, action *domain.Action, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	var actionName *string
	if action != nil {
		actionName = action.Name
	}

	var resourceName *string
	if resource != nil {
		resourceName = resource.Name
	}

	actionDefault := "*"
	if actionName != nil {
		actionDefault = *actionName
	}

	resourceDefault := "*"
	if resourceName != nil {
		resourceDefault = *resourceName
	}

	actionExpr := fmt.Sprintf("%s:%s", resourceDefault, actionDefault)

	if entitlement.ActionExpr != nil {
		if *entitlement.ActionExpr != actionExpr {
			vErrs.Append("action_expr", "action_expr is not valid")
		}
	}

	return nil
}

func (this *EntitlementServiceImpl) assertActionExprUnique(ctx crud.Context, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	dbEntitlement, err := this.entitlementRepo.FindByActionExpr(
		ctx,
		it.GetEntitlementByActionExprQuery{
			ActionExpr: *entitlement.ActionExpr,
			OrgId:      entitlement.OrgId,
		},
	)
	fault.PanicOnErr(err)

	if dbEntitlement != nil {
		vErrs.AppendAlreadyExists("action_expr", "action expression")
	}
	return nil
}

func (this *EntitlementServiceImpl) assertBusinessRuleCreateEntitlement(ctx crud.Context, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	resource, err := this.resourceService.GetResourceById(ctx, itResource.GetResourceByIdQuery{Id: *entitlement.ResourceId})
	fault.PanicOnErr(err)
	if resource.ClientError != nil {
		return resource.ClientError
	}

	action, err := this.actionService.GetActionById(ctx, itAction.GetActionByIdQuery{Id: *entitlement.ActionId})
	fault.PanicOnErr(err)
	if action.ClientError != nil {
		return action.ClientError
	}

	err = this.assertActionExprValid(resource.Data, action.Data, entitlement, vErrs)
	fault.PanicOnErr(err)

	err = this.assertEntitlementUnique(ctx, entitlement, vErrs)
	fault.PanicOnErr(err)

	err = this.assertActionExprUnique(ctx, entitlement, vErrs)
	fault.PanicOnErr(err)

	err = this.assertOrgExists(ctx, entitlement, vErrs)
	fault.PanicOnErr(err)

	return nil
}

func (this *EntitlementServiceImpl) assertOrgExists(ctx crud.Context, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	if entitlement.OrgId == nil {
		return nil
	}

	existCmd := &itOrg.ExistsOrgByIdCommand{
		Id: *entitlement.OrgId,
	}
	existRes := itOrg.ExistsOrgByIdResult{}
	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
	fault.PanicOnErr(err)

	if existRes.ClientError != nil {
		vErrs.MergeClientError(existRes.ClientError)
		return nil
	}

	if !existRes.Data {
		vErrs.Append("orgId", "not existing organization")
	}
	return nil
}

func (this *EntitlementServiceImpl) assertBusinessRuleDeleteEntitlement(ctx crud.Context, command it.DeleteEntitlementHardByIdCommand, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	_, err := this.assignmentService.DeleteByEntitlementId(ctx, itAssignment.DeleteEntitlementAssignmentByEntitlementIdCommand{EntitlementId: *entitlement.Id})
	fault.PanicOnErr(err)

	err = this.permissionHistoryRepo.EnableField(
		ctx,
		itPermissionHistory.EnableFieldCommand{
			EntitlementId:   entitlement.Id,
			EntitlementExpr: *entitlement.ActionExpr,
		},
	)
	fault.PanicOnErr(err)

	return nil
}
