package services

import (
	"strings"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// validateUomCreate enforces the invariants that apply to a brand-new UoM.
func validateUomCreate(repo models.UomSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		uom := models.NewUomFrom(params)
		assertFactorMatchesUomType(uom, vErrs)
		assertRoundingInRange(uom, vErrs)
		return assertSingleReferenceUom(ctx, repo, uom, nil, vErrs)
	}
}

// validateUomUpdate enforces the create-time invariants plus the historical-data-integrity
// rules that only apply once a record exists.
func validateUomUpdate(
	repo models.UomSearcher, usageDispatcher *usagecheck.Dispatcher,
) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		if foundModel == nil {
			return nil
		}
		params := inputModel.GetFieldData()
		uom, found := models.NewUomFrom(params), models.NewUomFrom(foundModel.GetFieldData())

		if err := assertImmutableWhileInUse(ctx, usageDispatcher, params, found, vErrs); err != nil {
			return err
		}
		// An update is partial: validate the resulting record, not just the submitted fields.
		merged := mergeUomForValidation(uom, found)
		assertFactorMatchesUomType(merged, vErrs)
		assertRoundingInRange(merged, vErrs)
		return assertSingleReferenceUom(ctx, repo, merged, found.GetId(), vErrs)
	}
}

// mergeUomForValidation overlays the submitted fields onto the stored record, so that a
// partial update is checked against the record it will produce.
func mergeUomForValidation(submitted *models.Uom, stored *models.Uom) *models.Uom {
	merged := dmodel.DynamicFields{}
	for key, val := range stored.GetFieldData() {
		merged[key] = val
	}
	for key, val := range submitted.GetFieldData() {
		merged[key] = val
	}
	return models.NewUomFrom(merged)
}

// assertFactorMatchesUomType enforces BR-UOM-ESS-006 and BR-UOM-ESS-009: the conversion
// factor must agree with the declared UoM type.
func assertFactorMatchesUomType(uom *models.Uom, vErrs *ft.ClientErrors) {
	uomType, factor := uom.GetUomType(), uom.GetFactor()
	if uomType == nil || factor == nil {
		// Absence is the schema's business, not ours.
		return
	}

	switch *uomType {
	case models.UomTypeReference:
		if !factor.Equal(decimal.NewFromInt(1)) {
			vErrs.Append(*ft.NewBusinessViolation(models.UomFieldFactor, "uom.reference_factor_must_be_one",
				"the Reference UoM must have a conversion factor of exactly 1"))
		}
	case models.UomTypeBiggerEqual:
		if factor.LessThan(decimal.NewFromInt(1)) {
			vErrs.Append(*ft.NewBusinessViolation(models.UomFieldFactor, "uom.bigger_equal_factor_out_of_range",
				"a bigger_equal UoM must have a conversion factor greater than or equal to 1"))
		}
	case models.UomTypeSmaller:
		if !factor.IsPositive() || !factor.LessThan(decimal.NewFromInt(1)) {
			vErrs.Append(*ft.NewBusinessViolation(models.UomFieldFactor, "uom.smaller_factor_out_of_range",
				"a smaller UoM must have a conversion factor greater than 0 and less than 1"))
		}
	}
}

// assertRoundingInRange enforces BR-UOM-ESS-017: 0 <= rounding <= 1. The upper bound is
// inclusive because a rounding step of exactly 1 is the meaningful "whole units only"
// precision of a discrete UoM such as Unit or gram (BR-UOM-ESS-015).
func assertRoundingInRange(uom *models.Uom, vErrs *ft.ClientErrors) {
	rounding := uom.GetRounding()
	if rounding == nil {
		return
	}
	if rounding.IsNegative() || rounding.GreaterThan(decimal.NewFromInt(1)) {
		vErrs.Append(*ft.NewBusinessViolation(models.UomFieldRounding, "uom.rounding_out_of_range",
			"rounding precision must be greater than or equal to 0 and less than or equal to 1"))
	}
}

// assertSingleReferenceUom enforces BR-UOM-ESS-005 and UOM-ESS-INV-09: a category holds at
// most one Reference UoM. selfId excludes the record being updated from the search.
func assertSingleReferenceUom(
	ctx corectx.Context,
	repo models.UomSearcher,
	uom *models.Uom,
	selfId *string,
	vErrs *ft.ClientErrors,
) error {
	uomType, categoryId := uom.GetUomType(), uom.GetCategoryId()
	if uomType == nil || *uomType != models.UomTypeReference || categoryId == nil {
		return nil
	}

	// Size 2: the record being updated may itself be the category's reference, and it must
	// not be mistaken for a conflicting one.
	existing, err := models.FindCategoryReferenceUoms(ctx, repo, *categoryId, 2)
	if err != nil {
		return errors.Wrap(err, "assertSingleReferenceUom")
	}

	for _, item := range existing {
		other := models.NewUomFrom(item)
		// The record being updated is already the category's reference; it does not
		// conflict with itself.
		if selfId != nil && other.GetId() != nil && *other.GetId() == *selfId {
			continue
		}
		vErrs.Append(*ft.NewBusinessViolation(models.UomFieldUomType, "uom.category_already_has_reference",
			"this UoM Category already has a Reference UoM"))
		return nil
	}
	return nil
}

// assertImmutableWhileInUse enforces BR-UOM-ESS-020 and UOM-ESS-INV-10. Changing the factor,
// type or category of a UoM already used by a transaction would reinterpret historical
// quantities; the supported remedy is to archive the UoM and create a replacement.
//
// An error from the check is reported as an error rather than swallowed: the alternative is
// permitting an edit on the strength of an answer nobody gave.
func assertImmutableWhileInUse(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher,
	params dmodel.DynamicFields, found *models.Uom, vErrs *ft.ClientErrors,
) error {
	// Nothing immutable was submitted, so there is no reason to ask anybody.
	immutableFields := make([]string, 0, 3)
	for _, field := range []string{models.UomFieldFactor, models.UomFieldUomType, models.UomFieldCategoryId} {
		if _, submitted := params[field]; submitted {
			immutableFields = append(immutableFields, field)
		}
	}
	if len(immutableFields) == 0 {
		return nil
	}

	inUse, by, err := isUomInUse(ctx, dispatcher, found)
	if err != nil {
		return err
	}
	if !inUse {
		return nil
	}

	for _, field := range immutableFields {
		vErrs.Append(*ft.NewBusinessViolation(field, "uom.immutable_while_in_use",
			"this UoM is already used by "+by+" transactions; "+
				"archive it and create a replacement instead"))
	}
	return nil
}

// isUomInUse reports whether any consuming module's records reference this UoM, and which.
//
// Essential cannot answer this from its own tables: the references live in Sales, Purchase,
// Inventory and Accounting, and importing them here would invert the dependency the module
// boundary exists to keep one way. The question goes out over the command bus to the modules
// Essential lists as its dependants, and each answers about its own data.
//
// A module that FAILS to answer counts as "in use". It knows only that it does not know, and
// permitting the edit on that basis could reinterpret quantities already recorded — the one
// outcome BR-UOM-ESS-020 exists to prevent. Refusing an edit that might have been fine is
// recoverable; silently changing what a historical document means is not.
//
// With no dependants the answer is false, which is correct for a binary that loads no consuming
// module: nothing can be referencing the unit.
func isUomInUse(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher, found *models.Uom,
) (bool, string, error) {
	if found == nil || found.GetId() == nil {
		return false, "", nil
	}

	orgId := ""
	if found.GetOrgId() != nil {
		orgId = string(*found.GetOrgId())
	}
	ref := usagecheck.ResourceRef{
		ResourceName: usagecheck.ResourceUom,
		Identifier:   usagecheck.NewIdentifier(string(*found.GetId()), orgId),
	}

	outcome, err := dispatcher.CheckUsage(ctx, modconstants.EssentialModuleName, []usagecheck.ResourceRef{ref})
	if err != nil {
		return false, "", err
	}
	if outcome.MayDelete() {
		return false, "", nil
	}
	return true, strings.Join(outcome.BlockingModules(), ", "), nil
}
