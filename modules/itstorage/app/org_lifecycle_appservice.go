package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

func NewOrgLifecycleAppService(
	spaceSvc itSpace.SpaceDomainService, fileSvc itFile.FileDomainService,
) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{spaceSvc: spaceSvc, fileSvc: fileSvc}
}

// OrgLifecycleAppServiceImpl gives a new org its storage space and, on cleanup, removes the space,
// every file in it, and the stored objects.
//
// The space is computed from the org rather than read from infra/defaults: its name and owner are
// the org's id, which no static defaults file can carry.
type OrgLifecycleAppServiceImpl struct {
	spaceSvc itSpace.SpaceDomainService
	fileSvc  itFile.FileDomainService
}

// InitOrg is idempotent, so a retried initialization does not trip the sibling-name unique key.
// The space hangs under the tenant root; a binary without tenants has no root, so it is a root.
func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	existing, err := this.spaceSvc.FindOrgSpace(ctx, orgId)
	if err != nil || existing != nil {
		return err
	}
	root, err := this.spaceSvc.FindTenantRootSpace(ctx)
	if err != nil {
		return err
	}
	row := dmodel.DynamicFields{
		models.SpaceFieldName:          orgId,
		models.SpaceFieldOwnerType:     c.OwnerTypeOrg,
		models.SpaceFieldOwnerId:       orgId,
		models.SpaceFieldManagedByType: c.ManagedByTypeSystem,
	}
	if root != nil {
		row[models.SpaceFieldParentId] = *root.GetId()
	}
	_, err = composable.CreateOne(ctx, this.spaceSvc, row)
	return err
}

// CleanUpOrg deletes the rows in one transaction and the objects after it commits; an object that
// cannot be removed is logged as an orphan rather than failing the cleanup.
func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	space, err := this.spaceSvc.FindOrgSpace(ctx, orgId)
	if err != nil || space == nil {
		return err
	}
	var objectKeys []string
	err = composable.InTransaction(ctx, this.spaceSvc.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		objectKeys, err = this.fileSvc.DeleteSpaceTree(tranxCtx, space)
		return err
	})
	if err != nil {
		return err
	}
	this.fileSvc.RemoveObjects(ctx, objectKeys)
	return nil
}
