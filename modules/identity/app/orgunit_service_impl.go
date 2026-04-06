package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itHier "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewOrgUnitServiceImpl(
	orgunitRepo2 itHier.OrgUnitRepository,
	userRepo2 itUser.UserRepository,
	orgSvc itOrg.OrganizationService,
	cqrsBus cqrs.CqrsBus,
) itHier.OrgUnitService {
	return &OrgUnitServiceImpl{
		cqrsBus:      cqrsBus,
		orgSvc:       orgSvc,
		orgunitRepo2: orgunitRepo2,
		userRepo2:    userRepo2,
	}
}

type OrgUnitServiceImpl struct {
	cqrsBus      cqrs.CqrsBus
	orgSvc       itOrg.OrganizationService
	orgunitRepo2 itHier.OrgUnitRepository
	userRepo2    itUser.UserRepository
}

func (this *OrgUnitServiceImpl) CreateOrgUnit(
	ctx corectx.Context, cmd itHier.CreateOrgUnitCommand,
) (*itHier.CreateOrgUnitResult, error) {
	return corecrud.Create(ctx, dyn.CreateParam[domain.OrganizationalUnit, *domain.OrganizationalUnit]{
		Action:         "create orgunit level",
		BaseRepoGetter: this.orgunitRepo2,
		Data:           cmd,
		ValidateExtra: func(ctx corectx.Context, unit *domain.OrganizationalUnit, vErrs *ft.ClientErrors) error {
			this.applyOrgUnitPathForCreate(ctx, unit, vErrs)
			return nil
		},
	})
}

func (this *OrgUnitServiceImpl) applyOrgUnitPathForCreate(
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

func (this *OrgUnitServiceImpl) appendOrgUnitPathFromParent(
	ctx corectx.Context, unit *domain.OrganizationalUnit, vErrs *ft.ClientErrors,
	parentId *model.Id, org, self string,
) {
	parentRes, err := this.orgunitRepo2.GetOne(ctx, dyn.RepoGetOneParam{
		Filter:  dmodel.DynamicFields{basemodel.FieldId: *parentId},
		Columns: []string{domain.OrgUnitFieldPath, domain.OrgUnitFieldOrgId},
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

func (this *OrgUnitServiceImpl) DeleteOrgUnit(
	ctx corectx.Context, cmd itHier.DeleteOrgUnitCommand,
) (*itHier.DeleteOrgUnitResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete orgunit level",
		DbRepoGetter: this.orgunitRepo2,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *OrgUnitServiceImpl) GetOrgUnit(
	ctx corectx.Context, query itHier.GetOrgUnitQuery,
) (*itHier.GetOrgUnitResult, error) {
	return corecrud.GetOne[domain.OrganizationalUnit](ctx, corecrud.GetOneParam{
		Action:       "get orgunit level",
		DbRepoGetter: this.orgunitRepo2,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *OrgUnitServiceImpl) SearchOrgUnits(
	ctx corectx.Context, query itHier.SearchOrgUnitsQuery,
) (*itHier.SearchOrgUnitsResult, error) {
	return corecrud.Search[domain.OrganizationalUnit](ctx, corecrud.SearchParam{
		Action:       "search orgunit levels",
		DbRepoGetter: this.orgunitRepo2,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *OrgUnitServiceImpl) OrgUnitExists(
	ctx corectx.Context, query itHier.OrgUnitExistsQuery,
) (*itHier.OrgUnitExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if orgunit level exists",
		DbRepoGetter: this.orgunitRepo2,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *OrgUnitServiceImpl) UpdateOrgUnit(
	ctx corectx.Context, cmd itHier.UpdateOrgUnitCommand,
) (*itHier.UpdateOrgUnitResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Group, *domain.Group]{
		Action:       "update group",
		DbRepoGetter: this.orgunitRepo2,
		Data:         cmd,
	})
}

func (this *OrgUnitServiceImpl) ManageOrgUnitUsers(
	ctx corectx.Context, cmd itHier.ManageOrgUnitUsersCommand,
) (result *itHier.ManageOrgUnitUsersResult, err error) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage orgunit level users",
		DbRepoGetter:       this.orgunitRepo2,
		DestSchemaName:     domain.UserSchemaName,
		SrcId:              cmd.OrgUnitId,
		SrcIdFieldForError: "org_unit_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
	})
}
