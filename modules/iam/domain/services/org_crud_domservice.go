package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

// maxUserOrgs bounds the org switcher's list; a user in more organizations than this is an
// installation the switcher was never designed for.
const maxUserOrgs = 500

var orgUsersM2mParam = composable.ManageM2mParam{
	Action:          "manage organization users",
	DestSchemaName:  domain.UserSchemaName,
	JunctionIdField: domain.OrgUsrRelFieldId,
}

// NewOrgCrudDomainService is handed the composable default by the iam_org onion.
func NewOrgCrudDomainService(
	base composable.CrudDomainService,
	settingsSvc itExt.OrgSettingsInitExtService,
	usageDispatcher *usagecheck.Dispatcher,
) itOrg.OrgCrudDomainService {
	return &OrgCrudDomainServiceImpl{
		CrudDomainService: base,
		settingsSvc:       settingsSvc,
		usageDispatcher:   usageDispatcher,
	}
}

type OrgCrudDomainServiceImpl struct {
	composable.CrudDomainService
	settingsSvc     itExt.OrgSettingsInitExtService
	usageDispatcher *usagecheck.Dispatcher
}

// Create creates the organization, seeds its settings and makes its creator a member in one
// transaction, so an org never exists without the settings rows the application reads from it, and
// its creator can act in it straight away.
func (this *OrgCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd itOrg.CreateOrgCommand, options ...composable.CreateOptions,
) (*itOrg.CreateOrgResult, error) {
	var result *itOrg.CreateOrgResult
	err := composable.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		result, err = this.CrudDomainService.Create(tranxCtx, asProvisioning(cmd), options...)
		if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		orgId := composable.ParamId(result.Data)
		if err := this.initOrgSettings(tranxCtx, orgId); err != nil {
			return err
		}
		return this.addCreatorAsMember(tranxCtx, orgId)
	})
	if err != nil {
		return nil, errors.Wrap(err, "create organization")
	}
	return result, nil
}

func (this *OrgCrudDomainServiceImpl) initOrgSettings(ctx corectx.Context, orgId model.Id) error {
	if orgId == "" {
		return errors.New("the created organization has no id")
	}
	initResult, err := this.settingsSvc.InitOrgSettings(ctx, itExt.InitOwnerSettingsCommand{OwnerId: orgId})
	if err != nil {
		return err
	}
	// A rejected seeding is a defect in this call, not something the caller can correct.
	if initResult.ClientErrors.Count() > 0 {
		return initResult.ClientErrors.ToError()
	}
	return nil
}

// addCreatorAsMember skips a caller that is not a user: a service or an unauthenticated
// execution cannot be a member. A refused add is a defect, not a client mistake, so it rolls the
// create back.
func (this *OrgCrudDomainServiceImpl) addCreatorAsMember(ctx corectx.Context, orgId model.Id) error {
	perms := ctx.GetPermissions()
	if perms.Principal.Kind != corectx.PrincipalKindUser || perms.UserId == "" {
		return nil
	}
	result, err := composable.ManageM2m(ctx, this.Repository(), dmodel.DynamicFields{
		basemodel.FieldId:      orgId,
		composable.FieldM2mAdd: []model.Id{perms.UserId},
	}, orgUsersM2mParam)
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "add the organization creator as a member")
	}
	return nil
}

// Delete removes an organization the org lifecycle has marked deleted and cleaned up. It still
// refuses one another module references: every dependant answers for its own tables, and one that
// cannot answer refuses the delete.
func (this *OrgCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd itOrg.DeleteOrgCommand, options ...composable.DeleteOptions,
) (*itOrg.DeleteOrgResult, error) {
	opts := composable.DeleteOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(
		assertDeletedBeforeRemoval, this.assertDeletable, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func (this *OrgCrudDomainServiceImpl) assertDeletable(
	ctx corectx.Context, keys *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	return AssertOrgDeletable(ctx, this.usageDispatcher, string(composable.ParamId(keys.GetFieldData())), vErrs)
}

// ManageUsers links users to and unlinks them from the organization. Membership is read at
// request time, so no permission cache needs rebuilding.
func (this *OrgCrudDomainServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itOrg.ManageOrgUsersCommand,
) (*itOrg.ManageOrgUsersResult, error) {
	return composable.ManageM2m(ctx, this.Repository(), cmd, orgUsersM2mParam)
}

func (this *OrgCrudDomainServiceImpl) GetUserOrgs(
	ctx corectx.Context, userId model.Id,
) (*itOrg.GetUserOrgsResult, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(domain.OrgEdgeUsers, dmodel.Linked, userId),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
		*dmodel.NewSearchNode().NewCondition(domain.OrgFieldStatus, dmodel.Equals, string(domain.OrgStatusReady)),
	)
	graph.OrderBy(domain.OrgFieldDisplayName, dmodel.Asc)

	return this.Search(ctx, dmodel.DynamicFields{
		"fields": []string{basemodel.FieldId, domain.OrgFieldDisplayName, domain.OrgFieldSlug},
		"graph":  graph,
		"page":   0,
		"size":   maxUserOrgs,
	})
}
