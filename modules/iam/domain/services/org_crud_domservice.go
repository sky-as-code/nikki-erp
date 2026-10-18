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

// Create creates the organization and seeds its settings in one transaction, so an org never
// exists without the settings rows the application reads from it.
func (this *OrgCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd itOrg.CreateOrgCommand, options ...composable.CreateOptions,
) (*itOrg.CreateOrgResult, error) {
	var result *itOrg.CreateOrgResult
	err := composable.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		result, err = this.CrudDomainService.Create(tranxCtx, cmd, options...)
		if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		return this.initOrgSettings(tranxCtx, composable.ParamId(result.Data))
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

// Delete refuses an organization another module still references. Every dependant answers for
// its own tables; one that cannot answer refuses the delete.
func (this *OrgCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd itOrg.DeleteOrgCommand, options ...composable.DeleteOptions,
) (*itOrg.DeleteOrgResult, error) {
	opts := composable.DeleteOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(this.assertDeletable, opts.ValidateExtra)
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
	return composable.ManageM2m(ctx, this.Repository(), cmd, composable.ManageM2mParam{
		Action:          "manage organization users",
		DestSchemaName:  domain.UserSchemaName,
		JunctionIdField: domain.OrgUsrRelFieldId,
	})
}

func (this *OrgCrudDomainServiceImpl) GetUserOrgs(
	ctx corectx.Context, userId model.Id,
) (*itOrg.GetUserOrgsResult, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(domain.OrgEdgeUsers, dmodel.Linked, userId),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
	)
	graph.OrderBy(domain.OrgFieldDisplayName, dmodel.Asc)

	return this.Search(ctx, dmodel.DynamicFields{
		"fields": []string{basemodel.FieldId, domain.OrgFieldDisplayName, domain.OrgFieldSlug},
		"graph":  graph,
		"page":   0,
		"size":   maxUserOrgs,
	})
}
