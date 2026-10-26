package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

// NewSpaceDomainService is handed the composable default by the itstorage_space onion.
func NewSpaceDomainService(base composable.CrudDomainService) itSpace.SpaceDomainService {
	return &SpaceDomainServiceImpl{
		CrudDomainService: base,
		repo:              base.Repository().(itSpace.SpaceRepository),
	}
}

// SpaceDomainServiceImpl keeps the materialised path true: a create derives it from the parent,
// a rename rewrites it for the space and every descendant in one transaction.
type SpaceDomainServiceImpl struct {
	composable.CrudDomainService
	repo itSpace.SpaceRepository
}

func (this *SpaceDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	cErrs, err := this.derivePath(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: cErrs}, nil
	}
	// A new space holds nothing; the size is the roll-up of what is uploaded into it later.
	cmd[models.SpaceFieldSizeBytes] = int64(0)
	return this.CrudDomainService.Create(ctx, cmd, options...)
}

func (this *SpaceDomainServiceImpl) derivePath(ctx corectx.Context, cmd composable.CreateCommand) (ft.ClientErrors, error) {
	cErrs := ft.ClientErrors{}
	parentPath := ""
	if parentId := composable.ParamString(cmd, models.SpaceFieldParentId); parentId != "" {
		parent, err := this.FindSpace(ctx, parentId)
		if err != nil {
			return cErrs, err
		}
		if parent == nil {
			cErrs.Append(*ft.NewNotFoundError(models.SpaceFieldParentId))
			return cErrs, nil
		}
		parentPath = parent.GetPath()
	}
	cmd[models.SpaceFieldPath] = models.JoinPath(parentPath, composable.ParamString(cmd, models.SpaceFieldName))
	return cErrs, nil
}

// Update recomputes the path only on a rename. The path is never taken from the caller.
func (this *SpaceDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	delete(cmd, models.SpaceFieldPath)
	delete(cmd, models.SpaceFieldSizeBytes)
	newName := composable.ParamString(cmd, models.SpaceFieldName)
	if newName == "" {
		return this.CrudDomainService.Update(ctx, cmd, options...)
	}

	var result *composable.MutateResult
	cErrs, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var refused *ft.ClientErrors
			var err error
			result, refused, err = this.rename(tranxCtx, cmd, newName, options)
			return refused, err
		})
	if err != nil {
		return nil, errors.Wrap(err, "SpaceDomainService.Update")
	}
	if cErrs != nil && cErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: *cErrs}, nil
	}
	return result, nil
}

func (this *SpaceDomainServiceImpl) rename(
	ctx corectx.Context, cmd composable.UpdateCommand, newName string, options []composable.UpdateOptions,
) (*composable.MutateResult, *ft.ClientErrors, error) {
	space, err := this.FindSpace(ctx, composable.ParamId(cmd))
	if err != nil {
		return nil, nil, err
	}
	if space == nil {
		return nil, &ft.ClientErrors{*ft.NewAnonymousNotFoundError()}, nil
	}
	oldPath := space.GetPath()
	newPath := models.JoinPath(parentPathOf(oldPath), newName)
	cmd[models.SpaceFieldPath] = newPath

	result, err := this.CrudDomainService.Update(ctx, cmd, options...)
	if err != nil {
		return nil, nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return result, &result.ClientErrors, nil
	}
	if newPath != oldPath {
		if _, err := this.repo.RewriteDescendantPaths(ctx, oldPath, newPath); err != nil {
			return nil, nil, err
		}
	}
	return result, nil, nil
}

func parentPathOf(path string) string {
	cut := strings.LastIndex(path, models.PathSeparator)
	if cut < 0 {
		return ""
	}
	return path[:cut]
}

func (this *SpaceDomainServiceImpl) FindSpace(ctx corectx.Context, spaceId model.Id) (*models.Space, error) {
	if spaceId == "" {
		return nil, nil
	}
	return this.findOne(ctx, dmodel.DynamicFields{models.SpaceFieldId: spaceId})
}

func (this *SpaceDomainServiceImpl) FindTenantRootSpace(ctx corectx.Context) (*models.Space, error) {
	tenantId := actingTenantId(ctx)
	if tenantId == "" {
		return nil, nil
	}
	return this.findSystemSpaceOf(ctx, c.OwnerTypeTenant, tenantId)
}

func (this *SpaceDomainServiceImpl) FindOrgSpace(ctx corectx.Context, orgId model.Id) (*models.Space, error) {
	return this.findSystemSpaceOf(ctx, c.OwnerTypeOrg, orgId)
}

func (this *SpaceDomainServiceImpl) AddSize(ctx corectx.Context, spaceId model.Id, delta int64) error {
	if delta == 0 {
		return nil
	}
	return this.repo.AddSize(ctx, spaceId, delta)
}

func (this *SpaceDomainServiceImpl) findOne(ctx corectx.Context, keys dmodel.DynamicFields) (*models.Space, error) {
	found, err := this.Repository().FindByKeys(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "find space")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "find space")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewSpaceFrom(found.Data), nil
}

// findSystemSpaceOf searches rather than FindByKeys: (owner_type, owner_id) is no unique key, because an owner
// may also hold spaces it manages itself. The system-managed one is unique by construction (seed, initOrg).
func (this *SpaceDomainServiceImpl) findSystemSpaceOf(
	ctx corectx.Context, ownerType string, ownerId model.Id,
) (*models.Space, error) {
	graph := dmodel.NewSearchGraph()
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.SpaceFieldOwnerType, dmodel.Equals, ownerType),
		*dmodel.NewSearchNode().NewCondition(models.SpaceFieldOwnerId, dmodel.Equals, string(ownerId)),
		*dmodel.NewSearchNode().NewCondition(models.SpaceFieldManagedByType, dmodel.Equals, c.ManagedByTypeSystem),
	)
	found, err := this.Repository().Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: 1})
	if err != nil {
		return nil, errors.Wrap(err, "find space")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "find space")
	}
	if len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewSpaceFrom(found.Data.Items[0]), nil
}
