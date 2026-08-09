package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// conversionScale is the working precision of intermediate results. BR-UOM-ESS-018 requires
// intermediate calculations to keep enough precision that values such as 1 Yard = 0.9144 m
// survive a chain of conversions; rounding happens once, at the end.
const conversionScale = 24

func NewUomConversionDomainServiceImpl() itUom.UomConversionDomainService {
	return &UomConversionDomainServiceImpl{}
}

type UomConversionDomainServiceImpl struct {
}

// Convert re-expresses a quantity from one UoM into another (BR-UOM-ESS-013).
//
//	Quantity in Target UoM = Quantity in Source UoM x Source Factor / Target Factor
//
// Both UoMs must belong to the same category; the category is the conversion boundary.
func (this *UomConversionDomainServiceImpl) Convert(
	ctx corectx.Context, query itUom.ConvertQuantityQuery,
) (*itUom.ConvertQuantityResult, error) {
	vErrs := &ft.ClientErrors{}

	source, target, err := this.loadConversionPair(ctx, query.SourceUomId, query.TargetUomId, vErrs)
	if err != nil {
		return nil, errors.Wrap(err, "convert quantity")
	}
	if vErrs.Count() > 0 {
		return &itUom.ConvertQuantityResult{ClientErrors: *vErrs}, nil
	}

	exact := query.Quantity.
		Mul(*source.GetFactor()).
		DivRound(*target.GetFactor(), conversionScale)

	return &itUom.ConvertQuantityResult{
		Data: itUom.ConvertQuantityResultData{
			Quantity:      applyRounding(exact, target.GetRounding()),
			ExactQuantity: exact,
		},
		HasData: true,
	}, nil
}

// ToReference normalizes a quantity into the Reference UoM of its own category
// (BR-UOM-ESS-007), the form business modules store stock in.
func (this *UomConversionDomainServiceImpl) ToReference(
	ctx corectx.Context, query itUom.ToReferenceQuery,
) (*itUom.ToReferenceResult, error) {
	vErrs := &ft.ClientErrors{}

	source, err := this.loadUom(ctx, query.SourceUomId, models.UomFieldId, vErrs)
	if err != nil {
		return nil, errors.Wrap(err, "convert to reference uom")
	}
	if vErrs.Count() > 0 {
		return &itUom.ToReferenceResult{ClientErrors: *vErrs}, nil
	}

	reference, err := this.loadCategoryReferenceUom(ctx, source, vErrs)
	if err != nil {
		return nil, errors.Wrap(err, "convert to reference uom")
	}
	if vErrs.Count() > 0 {
		return &itUom.ToReferenceResult{ClientErrors: *vErrs}, nil
	}

	// The reference UoM has factor 1 by invariant, so the division is a no-op; it is kept
	// explicit so the formula stays the one in BR-UOM-ESS-013.
	exact := query.Quantity.
		Mul(*source.GetFactor()).
		DivRound(*reference.GetFactor(), conversionScale)

	return &itUom.ToReferenceResult{
		Data: itUom.ToReferenceResultData{
			Quantity:       applyRounding(exact, reference.GetRounding()),
			ExactQuantity:  exact,
			ReferenceUomId: *reference.GetId(),
		},
		HasData: true,
	}, nil
}

// GetUom fetches a single UoM, so that a consuming module can validate a UoM reference it
// holds without reaching into Essential's repositories.
func (this *UomConversionDomainServiceImpl) GetUom(
	ctx corectx.Context, query itUom.GetUomQuery,
) (*itUom.GetUomResult, error) {
	vErrs := &ft.ClientErrors{}

	uom, err := this.loadUom(ctx, query.Id, models.UomFieldId, vErrs)
	if err != nil {
		return nil, errors.Wrap(err, "get uom")
	}
	if uom == nil {
		// Absence is a legitimate answer to a lookup, not a violation for the caller to
		// render; the caller decides what a missing UoM means in its own context.
		return &itUom.GetUomResult{HasData: false}, nil
	}

	return &itUom.GetUomResult{
		Data: itUom.GetUomResultData{
			Id:         *uom.GetId(),
			Symbol:     util.ValueOrZeroOf(uom.GetSymbol()),
			CategoryId: util.ValueOrZeroOf(uom.GetCategoryId()),
			IsArchived: util.ValueOrZeroOf(uom.IsArchived()),
		},
		HasData: true,
	}, nil
}

// loadConversionPair fetches both ends of a conversion and enforces the category boundary
// of BR-UOM-ESS-012 and the archive rule of BR-UOM-ESS-019.
func (this *UomConversionDomainServiceImpl) loadConversionPair(
	ctx corectx.Context, sourceId model.Id, targetId model.Id, vErrs *ft.ClientErrors,
) (*models.Uom, *models.Uom, error) {
	source, err := this.loadUom(ctx, sourceId, "source_uom_id", vErrs)
	if err != nil {
		return nil, nil, err
	}
	target, err := this.loadUom(ctx, targetId, "target_uom_id", vErrs)
	if err != nil {
		return nil, nil, err
	}
	if vErrs.Count() > 0 {
		return nil, nil, nil
	}

	assertSameCategory(source, target, vErrs)
	// An archived UoM stays usable as a source so historical quantities remain readable
	// (BR-UOM-ESS-020), but must not be the target of a new conversion (UOM-ESS-INV-11).
	assertNotArchivedTarget(target, vErrs)

	return source, target, nil
}

func (this *UomConversionDomainServiceImpl) loadUom(
	ctx corectx.Context, uomId model.Id, field string, vErrs *ft.ClientErrors,
) (*models.Uom, error) {
	engine, ok := dynamicresource.Registry().GetEngine(models.UomSchemaName)
	if !ok {
		return nil, errors.Errorf("loadUom: the '%s' engine is not registered", models.UomSchemaName)
	}

	found, err := engine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.UomFieldId: uomId},
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadUom")
	}
	if !found.HasData {
		vErrs.Append(*ft.NewBusinessViolation(field, "uom.not_found", "the UoM does not exist"))
		return nil, nil
	}

	uom := models.NewUomFrom(found.Data)
	if uom.GetFactor() == nil {
		vErrs.Append(*ft.NewBusinessViolation(field, "uom.missing_factor",
			"the UoM has no conversion factor"))
		return nil, nil
	}
	return uom, nil
}

func (this *UomConversionDomainServiceImpl) loadCategoryReferenceUom(
	ctx corectx.Context, source *models.Uom, vErrs *ft.ClientErrors,
) (*models.Uom, error) {
	// A UoM that is itself the reference needs no second lookup.
	if uomType := source.GetUomType(); uomType != nil && *uomType == models.UomTypeReference {
		return source, nil
	}

	categoryId := source.GetCategoryId()
	if categoryId == nil {
		vErrs.Append(*ft.NewBusinessViolation(models.UomFieldCategoryId, "uom.missing_category",
			"the UoM belongs to no UoM Category"))
		return nil, nil
	}

	engine, ok := dynamicresource.Registry().GetEngine(models.UomSchemaName)
	if !ok {
		return nil, errors.Errorf("loadCategoryReferenceUom: the '%s' engine is not registered",
			models.UomSchemaName)
	}

	found, err := models.FindCategoryReferenceUoms(ctx, engine.ResourceRepository(), *categoryId, 1)
	if err != nil {
		return nil, errors.Wrap(err, "loadCategoryReferenceUom")
	}
	if len(found) == 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.UomFieldCategoryId, "uom.category_has_no_reference",
			"the UoM Category has no Reference UoM"))
		return nil, nil
	}

	reference := models.NewUomFrom(found[0])
	if reference.GetFactor() == nil {
		vErrs.Append(*ft.NewBusinessViolation(models.UomFieldCategoryId, "uom.missing_factor",
			"the Reference UoM has no conversion factor"))
		return nil, nil
	}
	return reference, nil
}

// assertSameCategory enforces BR-UOM-ESS-011 and BR-UOM-ESS-012. Converting across
// categories is a caller mistake, so it surfaces as a business violation rather than an
// error.
func assertSameCategory(source *models.Uom, target *models.Uom, vErrs *ft.ClientErrors) {
	sourceCat, targetCat := source.GetCategoryId(), target.GetCategoryId()
	if sourceCat == nil || targetCat == nil || *sourceCat != *targetCat {
		vErrs.Append(*ft.NewBusinessViolation("target_uom_id", "uom.category_mismatch",
			"the source and target UoM must belong to the same UoM Category"))
	}
}

func assertNotArchivedTarget(target *models.Uom, vErrs *ft.ClientErrors) {
	if isArchived := target.IsArchived(); isArchived != nil && *isArchived {
		vErrs.Append(*ft.NewBusinessViolation("target_uom_id", "uom.target_archived",
			"an archived UoM cannot be the target of a conversion"))
	}
}

// applyRounding quantizes a quantity to a UoM's rounding precision (BR-UOM-ESS-015). The
// precision is a step, not a number of digits: a rounding of 0.25 snaps to quarters. A zero
// or absent rounding means no quantization, keeping the exact value.
func applyRounding(quantity decimal.Decimal, rounding *decimal.Decimal) decimal.Decimal {
	if rounding == nil || !rounding.IsPositive() {
		return quantity
	}
	return quantity.DivRound(*rounding, 0).Mul(*rounding)
}
