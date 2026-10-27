package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// NewOrgUnitCrudDomainService is handed the composable default by the iam_orgunit onion.
func NewOrgUnitCrudDomainService(
	base composable.CrudDomainService, userRepo itUser.UserCrudRepository,
) itOrgUnit.OrgUnitCrudDomainService {
	return &OrgUnitCrudDomainServiceImpl{CrudDomainService: base, userRepo: userRepo}
}

type OrgUnitCrudDomainServiceImpl struct {
	composable.CrudDomainService
	userRepo itUser.UserCrudRepository
}

// Create derives the unit's path: [org_id, ...ancestor ids, own id]. The path is what a
// unit-scoped grant is matched against, so it is computed here and never accepted from a client.
func (this *OrgUnitCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd itOrgUnit.CreateOrgUnitCommand, options ...composable.CreateOptions,
) (*itOrgUnit.CreateOrgUnitResult, error) {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.applyPath)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *OrgUnitCrudDomainServiceImpl) applyPath(
	ctx corectx.Context, input *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	unit := domain.NewOrganizationalUnitFrom(input.GetFieldData())
	selfId, orgId := unit.GetId(), unit.GetOrgId()
	if selfId == nil || orgId == nil {
		return nil
	}
	parentId := unit.GetParentId()
	if parentId == nil || *parentId == "" {
		unit.SetPath([]string{string(*orgId), string(*selfId)})
		return nil
	}

	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{basemodel.FieldId: string(*parentId)},
		Fields: []string{domain.OrgUnitFieldPath, domain.OrgUnitFieldOrgId},
	})
	if err != nil {
		return errors.Wrap(err, "applyPath")
	}
	if !found.HasData {
		vErrs.Append(*ft.NewNotFoundError(domain.OrgUnitFieldParentId))
		return nil
	}
	parent := domain.NewOrganizationalUnitFrom(found.Data)
	if parentOrg := parent.GetOrgId(); parentOrg == nil || *parentOrg != *orgId {
		vErrs.Append(*ft.NewBusinessViolation(domain.OrgUnitFieldParentId, ft.ErrorKey("err_parent_org_mismatch", "iam"),
			"parent org unit must belong to the same organization"))
		return nil
	}
	unit.SetPath(append(append([]string{}, parent.GetPath()...), string(*selfId)))
	return nil
}

// ManageUsers moves users into the unit (add) or out of it (remove). A user belongs to at most
// one unit through iam_user.org_unit_id, so this writes that column rather than a junction. A
// removal only clears a user who is currently in this unit.
func (this *OrgUnitCrudDomainServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itOrgUnit.ManageOrgUnitUsersCommand,
) (*itOrgUnit.ManageOrgUnitUsersResult, error) {
	change, cErrs := composable.ParseM2mChange(cmd)
	if cErrs != nil {
		return &itOrgUnit.ManageOrgUnitUsersResult{ClientErrors: *cErrs}, nil
	}
	affected := 0
	refused, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			vErrs := ft.ClientErrors{}
			if err := this.assertMembershipChange(tranxCtx, *change, &vErrs); err != nil || vErrs.Count() > 0 {
				return &vErrs, err
			}
			var err error
			affected, err = this.applyMembership(tranxCtx, *change)
			return nil, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "manage org unit users")
	}
	if refused != nil {
		return &itOrgUnit.ManageOrgUnitUsersResult{ClientErrors: *refused}, nil
	}
	if affected == 0 {
		return &itOrgUnit.ManageOrgUnitUsersResult{}, nil
	}
	return &itOrgUnit.ManageOrgUnitUsersResult{
		Data:    dyn.MutateResultData{AffectedCount: affected, AffectedAt: model.NewModelDateTime()},
		HasData: true,
	}, nil
}

func (this *OrgUnitCrudDomainServiceImpl) assertMembershipChange(
	ctx corectx.Context, change composable.M2mChange, vErrs *ft.ClientErrors,
) error {
	unitExists, err := this.Repository().KeyExist(ctx, dmodel.DynamicFields{basemodel.FieldId: string(change.SrcId)})
	if err != nil {
		return err
	}
	if unitExists.ClientErrors.Count() > 0 || !unitExists.Data {
		vErrs.Append(*ft.NewNotFoundError(basemodel.FieldId))
		return nil
	}
	removed := map[model.Id]bool{}
	for _, userId := range change.Removed {
		removed[userId] = true
	}
	for _, userId := range change.Added {
		if removed[userId] {
			vErrs.Append(*ft.NewBusinessViolation(composable.FieldM2mAdd, ft.ErrorKey("err_overlapped_fields"),
				"a user cannot be added and removed in the same call", map[string]any{"fields": "add, remove"}))
			return nil
		}
		userExists, err := this.userRepo.KeyExist(ctx, dmodel.DynamicFields{basemodel.FieldId: string(userId)})
		if err != nil {
			return err
		}
		if userExists.ClientErrors.Count() > 0 || !userExists.Data {
			vErrs.Append(*ft.NewNotFoundError(composable.FieldM2mAdd))
			return nil
		}
	}
	// A service is a machine principal: it may join an org, never an org unit.
	return serviceTypeScope.refuseIdsOfType(ctx, this.userRepo, composable.FieldM2mAdd, change.Added, vErrs)
}

func (this *OrgUnitCrudDomainServiceImpl) applyMembership(ctx corectx.Context, change composable.M2mChange) (int, error) {
	affected := 0
	for _, userId := range change.Added {
		if err := this.setUserUnit(ctx, userId, string(change.SrcId)); err != nil {
			return 0, err
		}
		affected++
	}
	for _, userId := range change.Removed {
		found, err := this.userRepo.GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{basemodel.FieldId: string(userId)},
			Fields: []string{domain.UserFieldOrgUnitId},
		})
		if err != nil {
			return 0, err
		}
		if !found.HasData || composable.ParamString(found.Data, domain.UserFieldOrgUnitId) != string(change.SrcId) {
			continue
		}
		if err := this.setUserUnit(ctx, userId, nil); err != nil {
			return 0, err
		}
		affected++
	}
	return affected, nil
}

func (this *OrgUnitCrudDomainServiceImpl) setUserUnit(ctx corectx.Context, userId model.Id, unitId any) error {
	result, err := this.userRepo.Update(ctx, dmodel.DynamicFields{
		basemodel.FieldId:         string(userId),
		domain.UserFieldOrgUnitId: unitId,
	})
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return result.ClientErrors.ToError()
	}
	return nil
}
