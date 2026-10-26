package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

// DefaultPossibleCountCap bounds how far the planner counts possible complete combinations. Past it
// the count is reported as over the threshold instead of materialized or estimated.
const DefaultPossibleCountCap = 100000

// Translation keys of the planner's violations.
const (
	ReasonPlanDuplicateLine       = "product_template.duplicate_attribute_line"
	ReasonPlanNoValidCombination  = "product_template.no_valid_combination"
	ReasonPlanTooManyVariants     = "product_template.too_many_variants"
	ReasonPlanCombinationConflict = "product_template.combination_conflict"
	ReasonPlanRoleConflict        = "product_template.reconcile_conflict"
	ReasonSelectionUnknownAttr    = "product_template.selection_unknown_attribute"
	ReasonSelectionValueNotAllow  = "product_template.selection_value_not_allowed"
	ReasonSelectionMissingAttr    = "product_template.selection_missing_attribute"
	ReasonSelectionExcluded       = "product_template.selection_excluded"
)

// PlanFieldAttributeLines is the payload field planner violations point at; callers may rename it.
const PlanFieldAttributeLines = "attribute_lines"

// PlanLine is one enabled attribute line of a template: the catalogue attribute, its mode and the
// catalogue value ids that are selected and available on the template.
type PlanLine struct {
	AttributeId string
	Mode        models.VariantCreationMode
	ValueIds    []string
}

// PlanExclusion forbids two catalogue values from appearing in one combination. The pair is
// unordered.
type PlanExclusion struct {
	LeftValueId  string `json:"left_value_id"`
	RightValueId string `json:"right_value_id"`
}

// PlanConfig is a template's attribute configuration. MaxPreCreated <= 0 means maxTemplateVariants.
type PlanConfig struct {
	Lines         []PlanLine
	Exclusions    []PlanExclusion
	MaxPreCreated int
}

// PlanVariant is an existing variant record of the template, archived ones included.
type PlanVariant struct {
	Id            string
	Kind          models.VariantKind
	Key           string
	IsArchived    bool
	ArchiveSource models.ArchiveSource
	BaseVariantId string

	// Roles is the mode each attribute of Key had when the record was created (from the line
	// snapshot). Optional: when provided and different from the next configuration, an otherwise
	// restorable record is a conflict instead of a restore (CR02 §8.5).
	Roles map[string]models.VariantCreationMode
}

// PlanInput is everything the planner needs. Previous is the configuration the existing variants
// were reconciled against; nil means no generating attribute (a fresh or default-only template).
type PlanInput struct {
	Previous         *PlanConfig
	Next             PlanConfig
	Existing         []PlanVariant
	TemplateArchived bool
	CountCap         int
}

// PlanOutcome is what the executor must do with one variant record or target key.
type PlanOutcome string

const (
	PlanOutcomeKeep      = PlanOutcome("keep")
	PlanOutcomeExtend    = PlanOutcome("extend")
	PlanOutcomeRestore   = PlanOutcome("restore")
	PlanOutcomeCreate    = PlanOutcome("create")
	PlanOutcomeInvalid   = PlanOutcome("invalid")
	PlanOutcomeConflict  = PlanOutcome("conflict")
	PlanOutcomeUntouched = PlanOutcome("untouched")
)

// PlannedVariant is one line of the plan. VariantId is empty for a create. NewKey is set for an
// extend. BaseKey is the instant sub-key of a complete in a mixed template, so the executor can
// point base_variant_id at the base holding that key (possibly created by the same plan).
type PlannedVariant struct {
	VariantId     string               `json:"variant_id,omitempty"`
	Kind          models.VariantKind   `json:"variant_kind"`
	Key           string               `json:"combination_key"`
	NewKey        string               `json:"new_combination_key,omitempty"`
	BaseKey       string               `json:"base_combination_key,omitempty"`
	IsArchived    bool                 `json:"is_archived"`
	ArchiveSource models.ArchiveSource `json:"archive_source,omitempty"`
	Outcome       PlanOutcome          `json:"outcome"`
}

// VariantPlan is the preview-ready result. Outcome lists are ordered bases first, then completes.
// Invalid holds active records plus template_cascade-archived ones (which become system_sync);
// Untouched holds archived records the plan leaves alone. Possible is nil when over the threshold.
type VariantPlan struct {
	Keep      []PlannedVariant `json:"keep"`
	Extend    []PlannedVariant `json:"extend"`
	Restore   []PlannedVariant `json:"restore"`
	Create    []PlannedVariant `json:"create"`
	Invalid   []PlannedVariant `json:"invalid"`
	Conflict  []PlannedVariant `json:"conflict"`
	Untouched []PlannedVariant `json:"untouched"`

	PreCreateKind         models.VariantKind `json:"pre_create_kind,omitempty"`
	PreCreateCount        int                `json:"pre_create_count"`
	PossibleCompleteCount *int               `json:"possible_complete_count"`
	PossibleOverThreshold bool               `json:"possible_over_threshold"`
	InactiveExclusions    []PlanExclusion    `json:"inactive_exclusions"`

	ClientErrors ft.ClientErrors `json:"-"`
}

// HasViolations reports whether the plan must be refused as a whole.
func (this *VariantPlan) HasViolations() bool {
	return this.ClientErrors.Count() > 0
}

func (this *VariantPlan) add(item PlannedVariant) {
	switch item.Outcome {
	case PlanOutcomeKeep:
		this.Keep = append(this.Keep, item)
	case PlanOutcomeExtend:
		this.Extend = append(this.Extend, item)
	case PlanOutcomeRestore:
		this.Restore = append(this.Restore, item)
	case PlanOutcomeCreate:
		this.Create = append(this.Create, item)
	case PlanOutcomeInvalid:
		this.Invalid = append(this.Invalid, item)
	case PlanOutcomeConflict:
		this.Conflict = append(this.Conflict, item)
	default:
		this.Untouched = append(this.Untouched, item)
	}
}

func (this *VariantPlan) violate(key string, message string, vars map[string]any) {
	this.ClientErrors.Append(*ft.NewBusinessViolation(
		PlanFieldAttributeLines, ft.ErrorKey(key, "inventory"), message, vars))
}
