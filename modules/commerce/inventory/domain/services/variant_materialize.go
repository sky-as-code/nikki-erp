package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Writing one variant record of a template: the variant row and the join rows naming the template
// values its combination is built from. Shared by the reconcile executor and on-demand resolve,
// so a variant is never created without its join rows. Every helper runs inside the caller's
// transaction and never authorizes.

// variantFieldSourceSystem and variantSourceSystemManual stamp the variants this module creates
// itself.
const (
	variantFieldSourceSystem  = "source_system"
	variantSourceSystemManual = "manual"
)

// TemplateValueRef is the template value row a catalogue value is selected through, on the
// template's enabled line for that attribute.
type TemplateValueRef struct {
	TemplateValueId string
	AttributeId     string
	SalesPriceExtra *decimal.Decimal
}

// TemplateValueRefs maps catalogue value id -> template value row, for the values currently
// selected on enabled lines.
type TemplateValueRefs map[string]TemplateValueRef

// LoadTemplateValueRefs reads the values currently selected on a template's enabled lines.
func LoadTemplateValueRefs(ctx corectx.Context, templateId string) (TemplateValueRefs, error) {
	state, err := loadTemplateConfigState(ctx, templateId)
	if err != nil {
		return nil, err
	}
	return state.valueRefs(), nil
}

// VariantMaterializeSpec describes the variant to create. ValueRefs is optional: nil loads the
// template's current refs. BaseVariantId is set for a complete of a mixed template.
type VariantMaterializeSpec struct {
	TemplateId    string
	OrgId         string
	Kind          models.VariantKind
	Key           string
	BaseVariantId string
	ValueRefs     TemplateValueRefs
}

// MaterializeVariant creates the variant through the variant domain service, so schema defaults,
// audit fields and the combination rule apply, then writes one join row per pair of the key. A
// pair whose value is not selected on the template is a client error and nothing is written.
func MaterializeVariant(
	ctx corectx.Context, spec VariantMaterializeSpec,
) (string, ft.ClientErrors, error) {
	refs, err := ensureValueRefs(ctx, spec.TemplateId, spec.ValueRefs)
	if err != nil {
		return "", nil, err
	}
	tavRefs, cErrs := refsForKey(refs, spec.Key)
	if cErrs.Count() > 0 {
		return "", cErrs, nil
	}
	variantSvc, err := domainServiceFor(models.ProductVariantSchemaName)
	if err != nil {
		return "", nil, err
	}
	created, err := variantSvc.Create(ctx, newVariantFields(spec))
	if err != nil {
		return "", nil, errors.Wrap(err, "MaterializeVariant")
	}
	if created.ClientErrors.Count() > 0 || !created.HasData {
		return "", created.ClientErrors, nil
	}
	variantId := derefString(models.NewProductVariantFrom(created.Data).GetId())
	if cErrs, err := writeVariantValueRows(ctx, variantId, spec.OrgId, tavRefs); err != nil || cErrs.Count() > 0 {
		return "", cErrs, err
	}
	return variantId, nil, nil
}

func newVariantFields(spec VariantMaterializeSpec) dmodel.DynamicFields {
	kind := spec.Kind
	if kind == "" {
		kind = models.VariantKindComplete
	}
	fields := dmodel.DynamicFields{
		models.ProductVariantFieldProductTemplateId: spec.TemplateId,
		models.ProductVariantFieldCombinationKey:    spec.Key,
		models.ProductVariantFieldVariantKind:       kind.String(),
		models.ProductVariantFieldStatus:            models.ProductVariantStatusActive.String(),
		variantFieldSourceSystem:                    variantSourceSystemManual,
		models.ProductVariantFieldOrgId:             spec.OrgId,
	}
	if spec.BaseVariantId != "" {
		fields[models.ProductVariantFieldBaseVariantId] = spec.BaseVariantId
	}
	return fields
}

// ExtendVariantKey gives an existing variant its extended combination (CR01 §6.2, CR02 §7.2):
// the key is rewritten and a join row is added for every pair the old key lacked. The row is
// written through the repository: combination_key is not client-editable identity.
func ExtendVariantKey(
	ctx corectx.Context, variantId string, orgId string, oldKey string, newKey string, refs TemplateValueRefs,
) (ft.ClientErrors, error) {
	oldPairs := ParseKeyToPairs(oldKey)
	added := map[string]string{}
	for attributeId, valueId := range ParseKeyToPairs(newKey) {
		if oldPairs[attributeId] != valueId {
			added[attributeId] = valueId
		}
	}
	tavRefs, cErrs := refsForKey(refs, BuildKeyFromPairs(added))
	if cErrs.Count() > 0 {
		return cErrs, nil
	}
	variantRepo, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}
	updated, err := variantRepo.Update(ctx, dmodel.DynamicFields{
		models.ProductVariantFieldId: variantId, models.ProductVariantFieldCombinationKey: newKey,
	})
	if err != nil {
		return nil, errors.Wrap(err, "ExtendVariantKey")
	}
	if updated.ClientErrors.Count() > 0 {
		return updated.ClientErrors, nil
	}
	return writeVariantValueRows(ctx, variantId, orgId, tavRefs)
}

// RestoreSyncArchivedVariant lifts a system_sync archive. It does nothing for an active variant or
// one held by any other reason, which this path must not lift (plan D3).
func RestoreSyncArchivedVariant(ctx corectx.Context, variant *models.ProductVariant) (bool, error) {
	update, ok := LiftArchiveFields(variant, models.ArchiveSourceSystemSync)
	if !ok {
		return false, nil
	}
	return true, writeVariantFields(ctx, update, "RestoreSyncArchivedVariant")
}

// ArchiveVariantForConfiguration archives a variant the configuration no longer allows, raising
// an existing weaker reason to system_sync and leaving a stronger one alone.
func ArchiveVariantForConfiguration(ctx corectx.Context, variant *models.ProductVariant) (bool, error) {
	update, ok := RaiseArchiveFields(variant, models.ArchiveSourceSystemSync)
	if !ok {
		return false, nil
	}
	return true, writeVariantFields(ctx, update, "ArchiveVariantForConfiguration")
}

func writeVariantFields(ctx corectx.Context, update dmodel.DynamicFields, what string) error {
	variantRepo, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return err
	}
	result, err := variantRepo.Update(ctx, update)
	if err != nil {
		return errors.Wrap(err, what)
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), what)
	}
	return nil
}

func ensureValueRefs(ctx corectx.Context, templateId string, refs TemplateValueRefs) (TemplateValueRefs, error) {
	if refs != nil {
		return refs, nil
	}
	return LoadTemplateValueRefs(ctx, templateId)
}

// refsForKey resolves every pair of a key to its template value row. A value that is not selected
// on an enabled line of its own attribute is reported on combination_key.
func refsForKey(refs TemplateValueRefs, key string) ([]TemplateValueRef, ft.ClientErrors) {
	pairs := ParseCombinationKey(key)
	out := make([]TemplateValueRef, 0, len(pairs))
	cErrs := ft.ClientErrors{}
	for _, pair := range pairs {
		ref, found := refs[pair.ValueId]
		if !found || ref.AttributeId != pair.AttributeId {
			cErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldCombinationKey,
				ft.ErrorKey("product_template.value_not_selected", "inventory"),
				"value {{value_id}} is not selected on the template for attribute {{attribute_id}}",
				map[string]any{"value_id": pair.ValueId, "attribute_id": pair.AttributeId}))
			continue
		}
		out = append(out, ref)
	}
	return out, cErrs
}

// writeVariantValueRows creates the join rows of a variant through the join resource's domain
// service, copying each template value's sales_price_extra as that resource requires.
func writeVariantValueRows(
	ctx corectx.Context, variantId string, orgId string, refs []TemplateValueRef,
) (ft.ClientErrors, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	joinSvc, err := domainServiceFor(models.ProductVariantAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}
	for _, ref := range refs {
		row := dmodel.DynamicFields{
			models.ProductVariantAttributeValueFieldProductVariantId:         variantId,
			models.ProductVariantAttributeValueFieldTemplateAttributeValueId: ref.TemplateValueId,
			models.ProductVariantAttributeValueFieldOrgId:                    orgId,
		}
		if ref.SalesPriceExtra != nil {
			row.SetDecimal(models.ProductVariantAttributeValueFieldSalesPriceExtra, ref.SalesPriceExtra)
		}
		created, err := joinSvc.Create(ctx, row)
		if err != nil {
			return nil, errors.Wrap(err, "writeVariantValueRows")
		}
		if created.ClientErrors.Count() > 0 {
			return created.ClientErrors, nil
		}
	}
	return nil, nil
}
