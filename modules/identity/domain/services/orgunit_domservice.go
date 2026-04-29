package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itHier "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewOrgUnitDomainServiceImpl(
	orgUnitRepo itHier.OrgUnitRepository,
	userRepo itUser.UserRepository,
	orgSvc itOrg.OrganizationDomainService,
	cqrsBus cqrs.CqrsBus,
) itHier.OrgUnitDomainService {
	return &OrgUnitDomainServiceImpl{
		cqrsBus:     cqrsBus,
		orgSvc:      orgSvc,
		orgUnitRepo: orgUnitRepo,
		userRepo:    userRepo,
	}
}

type OrgUnitDomainServiceImpl struct {
	cqrsBus     cqrs.CqrsBus
	orgSvc      itOrg.OrganizationDomainService
	orgUnitRepo itHier.OrgUnitRepository
	userRepo    itUser.UserRepository
}

func (this *OrgUnitDomainServiceImpl) CreateOrgUnit(
	ctx corectx.Context, cmd itHier.CreateOrgUnitCommand,
) (*itHier.CreateOrgUnitResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.OrganizationalUnit, *domain.OrganizationalUnit]{
		Action:         "create orgunit level",
		BaseRepoGetter: this.orgUnitRepo,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, inputModel *domain.OrganizationalUnit, vErrs *ft.ClientErrors) error {
			this.applyOrgUnitPathForCreate(ctx, inputModel, vErrs)
			return nil
		},
	})
}

func (this *OrgUnitDomainServiceImpl) applyOrgUnitPathForCreate(
	ctx corectx.Context, unit *domain.OrganizationalUnit, vErrs *ft.ClientErrors,
) {
	selfId := string(*unit.GetId())
	orgId := string(*unit.GetOrgId())
	parentId := unit.GetParentId()
	if parentId == nil || *parentId == "" {
		unit.SetPath([]string{orgId, selfId})
		return
	}
	this.appendOrgUnitPathFromParent(ctx, unit, vErrs, parentId, orgId, selfId)
}

func (this *OrgUnitDomainServiceImpl) appendOrgUnitPathFromParent(
	ctx corectx.Context, unit *domain.OrganizationalUnit, vErrs *ft.ClientErrors,
	parentId *model.Id, org, self string,
) {
	parentRes, err := this.orgUnitRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{basemodel.FieldId: *parentId},
		Fields: []string{domain.OrgUnitFieldPath, domain.OrgUnitFieldOrgId},
	})
	if err != nil {
		vErrs.Append(*ft.NewAnonymousValidationError("err_internal", err.Error(), nil))
		return
	}
	if !parentRes.HasData {
		vErrs.Append(*ft.NewValidationError(domain.OrgUnitFieldParentId, "err_not_found",
			"parent org unit was not found", nil))
		return
	}
	parent := parentRes.Data
	pOrg := parent.GetOrgId()
	if pOrg == nil || string(*pOrg) != org {
		vErrs.Append(*ft.NewBusinessViolation(domain.OrgUnitFieldParentId, "err_parent_org_mismatch",
			"parent org unit must belong to the same organization", nil))
		return
	}
	base := parent.GetPath()
	next := make([]string, len(base)+1)
	copy(next, base)
	next[len(base)] = self
	unit.SetPath(next)
}

func (this *OrgUnitDomainServiceImpl) DeleteOrgUnit(
	ctx corectx.Context, cmd itHier.DeleteOrgUnitCommand,
) (*itHier.DeleteOrgUnitResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete orgunit level",
		DbRepoGetter: this.orgUnitRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *OrgUnitDomainServiceImpl) GetOrgUnit(
	ctx corectx.Context, query itHier.GetOrgUnitQuery,
) (*dyn.OpResult[domain.OrganizationalUnit], error) {
	return corecrud.GetOne[domain.OrganizationalUnit](ctx, corecrud.GetOneParam{
		Action:       "get orgunit level",
		DbRepoGetter: this.orgUnitRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *OrgUnitDomainServiceImpl) SearchOrgUnits(
	ctx corectx.Context, query itHier.SearchOrgUnitsQuery,
) (*itHier.SearchOrgUnitsResult, error) {
	return corecrud.Search[domain.OrganizationalUnit](ctx, corecrud.SearchParam{
		Action:       "search orgunit levels",
		DbRepoGetter: this.orgUnitRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *OrgUnitDomainServiceImpl) OrgUnitExists(
	ctx corectx.Context, query itHier.OrgUnitExistsQuery,
) (*itHier.OrgUnitExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if orgunit level exists",
		DbRepoGetter: this.orgUnitRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *OrgUnitDomainServiceImpl) UpdateOrgUnit(
	ctx corectx.Context, cmd itHier.UpdateOrgUnitCommand,
) (*itHier.UpdateOrgUnitResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Group, *domain.Group]{
		Action:       "update group",
		DbRepoGetter: this.orgUnitRepo,
		Data:         cmd,
	})
}

func (this *OrgUnitDomainServiceImpl) ManageOrgUnitUsers(
	ctx corectx.Context, cmd itHier.ManageOrgUnitUsersCommand,
) (result *itHier.ManageOrgUnitUsersResult, err error) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage orgunit level users",
		DbRepoGetter:       this.orgUnitRepo,
		DestSchemaName:     domain.UserSchemaName,
		SrcId:              cmd.OrgUnitId,
		SrcIdFieldForError: "org_unit_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
	})
}
