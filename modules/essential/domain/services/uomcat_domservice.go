package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

// NewUomCatDomainService is handed the composable default by the UoM Category onion. It also
// takes the UoM repository, because a category's Reference UoM is a UoM row it must look at.
func NewUomCatDomainService(
	base composable.CrudDomainService, uomRepo itUom.UomRepository,
) itUomCat.UomCatDomainService {
	return &UomCatDomainServiceImpl{CrudDomainService: base, uomRepo: uomRepo}
}

type UomCatDomainServiceImpl struct {
	composable.CrudDomainService
	uomRepo itUom.UomRepository
}

func (this *UomCatDomainServiceImpl) Create(
	ctx corectx.Context, cmd itUomCat.CreateUomCatCommand, options ...composable.CreateOptions,
) (*itUomCat.CreateUomCatResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, validateUomCatReference(this.uomRepo, nil))
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *UomCatDomainServiceImpl) Update(
	ctx corectx.Context, cmd itUomCat.UpdateUomCatCommand, options ...composable.UpdateOptions,
) (*itUomCat.UpdateUomCatResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra,
		validateUomCatReference(this.uomRepo, assertReferenceUomStableWhileInUse))
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// Delete refuses to remove a category that still has units of measure. Without the guard the UoM
// foreign key refuses the statement and the failure reaches the client as a 500.
func (this *UomCatDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.DeleteOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	vErrs := ft.NewClientErrors()
	categoryId := readUomCatIdParam(params, models.UomCatFieldId)
	if err := AssertUomCatDeletable(ctx, this.uomRepo, categoryId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
}

// AssertUomCatDeletable blocks the delete once any UoM belongs to the category.
func AssertUomCatDeletable(
	ctx corectx.Context, repo models.UomSearcher, categoryId string, vErrs *ft.ClientErrors,
) error {
	if categoryId == "" {
		return nil
	}

	uoms, err := models.FindCategoryUoms(ctx, repo, categoryId, 1)
	if err != nil {
		return errors.Wrap(err, "AssertUomCatDeletable")
	}
	if len(uoms) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.UomCatFieldId, "uomcat.has_uoms",
			"this category still has units of measure; move or delete them first"))
	}
	return nil
}

// readUomCatIdParam reads a delete parameter as a plain string. model.Id is a string type, so the
// string case covers it too.
func readUomCatIdParam(params dmodel.DynamicFields, field string) string {
	val, ok := params[field]
	if !ok || val == nil {
		return ""
	}
	if typed, ok := val.(string); ok {
		return typed
	}
	return ""
}

type uomCatUpdateCheckFn func(
	ctx corectx.Context, params dmodel.DynamicFields, found *models.UomCat, vErrs *ft.ClientErrors,
)

// validateUomCatReference enforces BR-UOM-ESS-004 and UOM-ESS-INV-03: a category's Reference
// UoM must be a UoM of that same category. onUpdate carries the checks that only make sense
// once a stored record exists.
func validateUomCatReference(uomRepo itUom.UomRepository, onUpdate uomCatUpdateCheckFn) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		if onUpdate != nil && foundModel != nil {
			onUpdate(ctx, params, models.NewUomCatFrom(foundModel.GetFieldData()), vErrs)
		}

		cat := models.NewUomCatFrom(params)
		referenceUomId := cat.GetReferenceUomId()
		if referenceUomId == nil {
			// A category may exist without a reference UoM until one is created for it.
			return nil
		}

		categoryId := cat.GetId()
		if categoryId == nil && foundModel != nil {
			categoryId = models.NewUomCatFrom(foundModel.GetFieldData()).GetId()
		}
		if categoryId == nil {
			// Create: the category has no id yet, so no existing UoM can already belong to
			// it. The reference must still exist (BR-UOM-ESS-004 needs something to point
			// at), but any UoM that does exist necessarily belongs to another category.
			return assertReferenceUomBelongsToCategory(ctx, uomRepo, *referenceUomId, "", vErrs)
		}
		return assertReferenceUomBelongsToCategory(ctx, uomRepo, *referenceUomId, *categoryId, vErrs)
	}
}

// assertReferenceUomBelongsToCategory reports the reference UoM as not-found when it does
// not exist, and as foreign when it belongs to a category other than categoryId. An empty
// categoryId is the create case: the category has no id yet, so every existing UoM is foreign.
func assertReferenceUomBelongsToCategory(
	ctx corectx.Context, uomRepo itUom.UomRepository, referenceUomId string, categoryId string, vErrs *ft.ClientErrors,
) error {
	found, err := uomRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.UomFieldId: referenceUomId},
	})
	if err != nil {
		return errors.Wrap(err, "assertReferenceUomBelongsToCategory")
	}
	if !found.HasData {
		vErrs.Append(*ft.NewBusinessViolation(models.UomCatFieldReferenceUomId, "uomcat.reference_uom_not_found",
			"the referenced UoM does not exist"))
		return nil
	}

	uom := models.NewUomFrom(found.Data)
	if uom.GetCategoryId() == nil || *uom.GetCategoryId() != categoryId {
		vErrs.Append(*ft.NewBusinessViolation(models.UomCatFieldReferenceUomId, "uomcat.reference_uom_foreign",
			"the Reference UoM must belong to this UoM Category"))
	}
	return nil
}

// assertReferenceUomStableWhileInUse enforces BR-UOM-ESS-021: once a category is in use,
// repointing its Reference UoM would reinterpret every factor in the category.
func assertReferenceUomStableWhileInUse(
	ctx corectx.Context, params dmodel.DynamicFields, found *models.UomCat, vErrs *ft.ClientErrors,
) {
	submitted, isSubmitted := params[models.UomCatFieldReferenceUomId]
	if !isSubmitted || !isUomCatInUse(ctx, found) {
		return
	}
	// Re-submitting the value unchanged is not a change.
	current := found.GetReferenceUomId()
	if current != nil && submitted == *current {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(models.UomCatFieldReferenceUomId, "uomcat.reference_uom_immutable_while_in_use",
		"this UoM Category is already in use; its Reference UoM can no longer be changed"))
}

// isUomCatInUse reports whether any product or transaction depends on this category.
//
// TODO: shares the fate of isUomInUse — no module consumes UoM yet. Wire a real probe when
// stock, purchase or sales land, or BR-UOM-ESS-021 stays structurally declared but inert.
func isUomCatInUse(_ corectx.Context, _ *models.UomCat) bool {
	return false
}
