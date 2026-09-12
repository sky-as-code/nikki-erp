package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

// NewUomCatDomainService is handed the composable default by the UoM Category onion. It also
// takes the UoM repository, because a category's Reference UoM is a UoM row it must look at.
func NewUomCatDomainService(
	base composable.CrudDomainService, uomRepo itUom.UomRepository,
	usageDispatcher *usagecheck.Dispatcher,
) itUomCat.UomCatDomainService {
	return &UomCatDomainServiceImpl{
		CrudDomainService: base,
		uomRepo:           uomRepo,
		usageDispatcher:   usageDispatcher,
	}
}

type UomCatDomainServiceImpl struct {
	composable.CrudDomainService
	uomRepo         itUom.UomRepository
	usageDispatcher *usagecheck.Dispatcher
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
		validateUomCatReference(this.uomRepo,
			assertReferenceUomStableWhileInUse(this.usageDispatcher, this.uomRepo)))
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
//
// A closure over the dispatcher and the UoM repository, because answering "is this category in
// use" now means asking the consuming modules about the units it holds.
func assertReferenceUomStableWhileInUse(
	dispatcher *usagecheck.Dispatcher, uomRepo itUom.UomRepository,
) uomCatUpdateCheckFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, found *models.UomCat, vErrs *ft.ClientErrors,
	) {
		assertReferenceUomStable(ctx, dispatcher, uomRepo, params, found, vErrs)
	}
}

func assertReferenceUomStable(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher, uomRepo itUom.UomRepository,
	params dmodel.DynamicFields, found *models.UomCat, vErrs *ft.ClientErrors,
) {
	submitted, isSubmitted := params[models.UomCatFieldReferenceUomId]
	if !isSubmitted || !isUomCatInUse(ctx, dispatcher, uomRepo, found) {
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

// isUomCatInUse reports whether anything depends on this category.
//
// No module references a category directly — they reference the UNITS in it — so the question is
// asked one level down: if any unit of this category is in use by a consuming module, the
// category is in use. Repointing the reference unit would then rescale every factor in the
// category, reinterpreting every quantity already recorded in any of its units.
//
// A check that cannot be completed counts as "in use", for the same reason it does on the unit
// itself: refusing an edit that would have been fine is recoverable, and silently rescaling
// history is not.
func isUomCatInUse(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher,
	repo models.UomSearcher, found *models.UomCat,
) bool {
	if found == nil || found.GetId() == nil {
		return false
	}
	categoryId := string(*found.GetId())

	// MaxCategoryUomsToCheck bounds the fan-out: a category with more units than this is asked
	// about only the first few, and the rest are assumed in use rather than left unchecked.
	uoms, err := models.FindCategoryUoms(ctx, repo, categoryId, maxCategoryUomsToCheck)
	if err != nil {
		return true
	}
	if len(uoms) == 0 {
		return false
	}
	if len(uoms) >= maxCategoryUomsToCheck {
		return true
	}

	orgId := ""
	if found.GetOrgId() != nil {
		orgId = string(*found.GetOrgId())
	}

	refs := make([]usagecheck.ResourceRef, 0, len(uoms))
	for _, uom := range uoms {
		uomId := models.NewUomFrom(uom).GetId()
		if uomId == nil {
			continue
		}
		refs = append(refs, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceUom,
			Identifier:   usagecheck.NewIdentifier(string(*uomId), orgId),
		})
	}
	if len(refs) == 0 {
		return false
	}

	// One dispatch carrying every unit, rather than one per unit: the command takes a list
	// precisely so a question about several resources costs one round trip per module.
	outcome, err := dispatcher.CheckUsage(ctx, modconstants.EssentialModuleName, refs)
	if err != nil {
		return true
	}
	return !outcome.MayDelete()
}

// maxCategoryUomsToCheck caps how many units one category asks about. A category holding more
// than this is certainly established enough that its reference unit should not be repointed.
const maxCategoryUomsToCheck = 20
