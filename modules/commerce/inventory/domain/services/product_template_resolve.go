package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Translation keys of the resolve refusals.
const (
	ReasonResolveTemplateArchived = "product_template.archived"
	ReasonResolveBaseMissing      = "product_variant.base_missing"
	ReasonResolveBaseArchived     = "product_variant.base_archived"
	ReasonResolveManuallyArchived = "product_variant.manually_archived"
)

var _ itProduct.CompleteVariantResolver = (*ProductTemplateDomainServiceImpl)(nil)

// resolveAction is what resolving must do once the selection is known to be valid.
type resolveAction string

const (
	resolveActionExisting = resolveAction("existing")
	resolveActionRestore  = resolveAction("restore")
	resolveActionCreate   = resolveAction("create")
	resolveActionMissing  = resolveAction("missing")
)

// resolveFacts are the records a valid selection maps to. Base is consulted only when Mixed.
type resolveFacts struct {
	Mixed       bool
	Base        *models.ProductVariant
	Complete    *models.ProductVariant
	Materialize bool
}

// ResolveCompleteVariant turns a full selection into its complete variant (CR01 §5.3, CR02 §5.2).
// Reads never write: only MaterializeIfMissing lets it create the missing complete or restore one
// a configuration change archived. A manually archived complete is never duplicated.
func (this *ProductTemplateDomainServiceImpl) ResolveCompleteVariant(
	ctx corectx.Context, query itProduct.ResolveCompleteVariantQuery,
) (*itProduct.ResolveCompleteVariantResult, error) {
	template, err := this.fetchTemplate(ctx, query.TemplateId)
	if err != nil || template == nil {
		return &itProduct.ResolveCompleteVariantResult{}, err
	}
	if refused := assertTemplateResolvable(template, query.ExpectedEtag); refused != nil {
		return &itProduct.ResolveCompleteVariantResult{ClientErrors: *refused}, nil
	}

	state, err := loadTemplateConfigState(ctx, query.TemplateId)
	if err != nil {
		return nil, err
	}
	config := state.planConfig()
	completeKey, baseKey, violations := ValidateSelection(config, query.Selections)
	if violations.Count() > 0 {
		return &itProduct.ResolveCompleteVariantResult{ClientErrors: violations}, nil
	}

	facts, err := loadResolveFacts(ctx, query.TemplateId, completeKey, baseKey)
	if err != nil {
		return nil, err
	}
	facts.Materialize = query.MaterializeIfMissing
	data := itProduct.ResolveCompleteVariantResultData{
		CombinationKey:  completeKey,
		NeverSelections: neverSelections(config, query.Selections),
	}
	return this.applyResolve(ctx, template, facts, data, state.valueRefs())
}

// applyResolve carries out the decided action and fills the answer.
func (this *ProductTemplateDomainServiceImpl) applyResolve(
	ctx corectx.Context, template *models.ProductTemplate, facts resolveFacts,
	data itProduct.ResolveCompleteVariantResultData, refs TemplateValueRefs,
) (*itProduct.ResolveCompleteVariantResult, error) {
	action, refusal := decideResolve(facts)
	if refusal != nil {
		return &itProduct.ResolveCompleteVariantResult{ClientErrors: ft.ClientErrors{*refusal}}, nil
	}
	if facts.Mixed {
		data.BaseVariantId = derefString(facts.Base.GetId())
	}

	switch action {
	case resolveActionExisting:
		data.VariantId, data.Outcome = derefString(facts.Complete.GetId()), itProduct.ResolveOutcomeExisting
	case resolveActionMissing:
		data.Outcome = itProduct.ResolveOutcomeMissing
	case resolveActionRestore:
		if _, err := RestoreSyncArchivedVariant(ctx, facts.Complete); err != nil {
			return nil, err
		}
		data.VariantId, data.Outcome = derefString(facts.Complete.GetId()), itProduct.ResolveOutcomeRestored
	case resolveActionCreate:
		return this.createResolvedComplete(ctx, template, data, refs)
	}
	return &itProduct.ResolveCompleteVariantResult{Data: data, HasData: true}, nil
}

// decideResolve is the pure outcome rule. A mixed template needs its base active: a missing base
// means the configuration was never reconciled, and completing it would orphan the complete.
func decideResolve(facts resolveFacts) (resolveAction, *ft.ClientErrorItem) {
	if facts.Mixed {
		if facts.Base == nil {
			return "", resolveViolation(ReasonResolveBaseMissing,
				"the base variant for these values is missing; reconcile the product's configuration", nil)
		}
		if source := EffectiveArchiveSource(facts.Base); source != "" {
			return "", resolveViolation(ReasonResolveBaseArchived,
				"the base variant for these values is archived", map[string]any{"archive_source": source})
		}
	}
	if facts.Complete == nil {
		if facts.Materialize {
			return resolveActionCreate, nil
		}
		return resolveActionMissing, nil
	}
	switch EffectiveArchiveSource(facts.Complete) {
	case "":
		return resolveActionExisting, nil
	case models.ArchiveSourceSystemSync:
		if facts.Materialize {
			return resolveActionRestore, nil
		}
		return resolveActionMissing, nil
	}
	return "", resolveViolation(ReasonResolveManuallyArchived,
		"the variant for these values was archived; unarchive it instead of creating a new one",
		map[string]any{"variant_id": derefString(facts.Complete.GetId())})
}

// assertTemplateResolvable refuses an archived template and a stale configuration version.
func assertTemplateResolvable(template *models.ProductTemplate, expectedEtag *string) *ft.ClientErrors {
	if archived := template.IsArchived(); archived != nil && *archived {
		return &ft.ClientErrors{*resolveViolation(ReasonResolveTemplateArchived,
			"the product is archived; unarchive it first", nil)}
	}
	if expectedEtag == nil {
		return nil
	}
	if current := template.GetEtag(); current == nil || string(*current) != *expectedEtag {
		return &ft.ClientErrors{*ft.NewEtagMismatchedError()}
	}
	return nil
}

func resolveViolation(key string, message string, vars map[string]any) *ft.ClientErrorItem {
	if vars == nil {
		return ft.NewBusinessViolation(PlanFieldSelection, ft.ErrorKey(key, "inventory"), message)
	}
	return ft.NewBusinessViolation(PlanFieldSelection, ft.ErrorKey(key, "inventory"), message, vars)
}

// loadResolveFacts reads the base (mixed only) and the complete holding the keys, archived
// records included: an archived complete must be found so it is never duplicated.
func loadResolveFacts(
	ctx corectx.Context, templateId string, completeKey string, baseKey string,
) (resolveFacts, error) {
	facts := resolveFacts{Mixed: baseKey != ""}
	complete, err := findKeyedVariant(ctx, templateId, models.VariantKindComplete, completeKey)
	if err != nil {
		return facts, err
	}
	facts.Complete = complete
	if facts.Mixed {
		facts.Base, err = findKeyedVariant(ctx, templateId, models.VariantKindBase, baseKey)
	}
	return facts, err
}

// findKeyedVariant finds the one variant of a kind holding a key; (template, kind, key) is unique.
func findKeyedVariant(
	ctx corectx.Context, templateId string, kind models.VariantKind, key string,
) (*models.ProductVariant, error) {
	repo, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}
	rows, err := models.FindVariantsByCombination(ctx, repo, templateId, key, 2)
	if err != nil {
		return nil, errors.Wrap(err, "findKeyedVariant")
	}
	for _, row := range rows {
		variant := models.NewProductVariantFrom(row)
		if variantKindOf(variant) == kind {
			return variant, nil
		}
	}
	return nil, nil
}

// neverSelections echoes the never-line values of a validated selection, in attribute order.
func neverSelections(config PlanConfig, selection map[string]string) []itProduct.AttributeSelection {
	echoed := []itProduct.AttributeSelection{}
	for _, line := range config.Lines {
		valueId, found := selection[line.AttributeId]
		if found && !line.Mode.CreatesVariants() {
			echoed = append(echoed, itProduct.AttributeSelection{
				AttributeId: line.AttributeId, ValueId: valueId, Mode: line.Mode,
			})
		}
	}
	sort.Slice(echoed, func(i, j int) bool { return echoed[i].AttributeId < echoed[j].AttributeId })
	return echoed
}
