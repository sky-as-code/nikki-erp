package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductCategoryDomainService attaches the category-tree invariant to create and update. The
// CRUD processing is the default's; only the rule the schema cannot express belongs here.
func NewProductCategoryDomainService(base composable.CrudDomainService) itProduct.ProductCategoryDomainService {
	return &ProductCategoryDomainServiceImpl{CrudDomainService: base}
}

type ProductCategoryDomainServiceImpl struct {
	composable.CrudDomainService
}

func (this *ProductCategoryDomainServiceImpl) Create(
	ctx corectx.Context, cmd itProduct.CreateProductCategoryCommand, options ...composable.CreateOptions,
) (*itProduct.CreateProductCategoryResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateCategoryCreate)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *ProductCategoryDomainServiceImpl) Update(
	ctx corectx.Context, cmd itProduct.UpdateProductCategoryCommand, options ...composable.UpdateOptions,
) (*itProduct.UpdateProductCategoryResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateCategoryUpdate)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// validateCategoryCreate rejects a category that would be its own ancestor. On create the record
// has no id yet, so only the parent chain itself can be malformed.
func (this *ProductCategoryDomainServiceImpl) validateCategoryCreate(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	category := models.NewProductCategoryFrom(inputModel.GetFieldData())
	return this.assertNoCategoryCycle(ctx, derefString(category.GetParentCategoryId()), "", vErrs)
}

// validateCategoryUpdate additionally rejects re-parenting a category under one of its own
// descendants, which is how a cycle is normally introduced.
func (this *ProductCategoryDomainServiceImpl) validateCategoryUpdate(
	ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if foundModel == nil {
		return nil
	}
	submitted := models.NewProductCategoryFrom(inputModel.GetFieldData())
	parentId := submitted.GetParentCategoryId()
	if parentId == nil {
		// The parent is not being changed, so the existing chain still holds.
		return nil
	}

	stored := models.NewProductCategoryFrom(foundModel.GetFieldData())
	selfId := derefString(stored.GetId())

	if derefString(parentId) == selfId {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
			"product_category.self_parent",
			"a category cannot be its own parent"))
		return nil
	}
	return this.assertNoCategoryCycle(ctx, derefString(parentId), selfId, vErrs)
}

// assertNoCategoryCycle walks upwards from parentId. Reaching selfId means the proposed parent
// is a descendant of the category being edited, so the edit would close a loop.
func (this *ProductCategoryDomainServiceImpl) assertNoCategoryCycle(
	ctx corectx.Context, parentId string, selfId string, vErrs *ft.ClientErrors,
) error {
	if parentId == "" {
		return nil
	}

	seen := map[string]bool{}
	current := parentId
	for depth := 0; current != "" && depth < maxCategoryDepth; depth++ {
		if current == selfId {
			vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
				"product_category.cycle",
				"this parent is a descendant of the category, which would create a cycle"))
			return nil
		}
		if seen[current] {
			// A pre-existing cycle in stored data. Stop rather than loop, and let the edit
			// through: it is not the change under validation that introduced it.
			return nil
		}
		seen[current] = true

		found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{models.ProductCategoryFieldId: current},
			Fields: []string{models.ProductCategoryFieldId, models.ProductCategoryFieldParentCategoryId},
		})
		if err != nil {
			return errors.Wrap(err, "assertNoCategoryCycle")
		}
		if !found.HasData {
			// A missing parent is the reference check's business, not ours.
			return nil
		}
		current = derefString(models.NewProductCategoryFrom(found.Data).GetParentCategoryId())
	}
	return nil
}
