package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// ResolveOutcome tells the caller what resolving a selection did.
type ResolveOutcome string

const (
	// ResolveOutcomeExisting found an active complete variant holding the combination.
	ResolveOutcomeExisting = ResolveOutcome("existing")
	// ResolveOutcomeCreated created the complete variant, because none had ever held the key.
	ResolveOutcomeCreated = ResolveOutcome("created")
	// ResolveOutcomeRestored unarchived the variant a configuration change had archived.
	ResolveOutcomeRestored = ResolveOutcome("restored")
	// ResolveOutcomeMissing is a lookup-only answer: the combination is valid but no active
	// variant holds it, and the caller did not ask for one to be materialized.
	ResolveOutcomeMissing = ResolveOutcome("missing")
)

// ResolveCompleteVariantQuery is a full selection for one template, keyed by catalogue attribute
// id. The org was already confined by the application service.
type ResolveCompleteVariantQuery struct {
	TemplateId   string
	Selections   map[string]string
	ExpectedEtag *string

	// MaterializeIfMissing allows the write: creating the missing complete, or restoring one a
	// configuration change archived. False is a pure lookup that never writes.
	MaterializeIfMissing bool
}

type ResolveCompleteVariantResultData struct {
	VariantId      string
	BaseVariantId  string
	CombinationKey string
	Outcome        ResolveOutcome

	// NeverSelections are the values of never lines: validated and echoed for the caller to keep
	// on its own line, never part of the variant.
	NeverSelections []AttributeSelection
}

type ResolveCompleteVariantResult = dyn.OpResult[ResolveCompleteVariantResultData]

// CompleteVariantResolver turns a full selection into the complete variant a transaction line
// must reference, creating it on demand.
type CompleteVariantResolver interface {
	ResolveCompleteVariant(
		ctx corectx.Context, query ResolveCompleteVariantQuery,
	) (*ResolveCompleteVariantResult, error)
}

type ResolveSelectionPairView struct {
	AttributeId string `json:"attribute_id"`
	ValueId     string `json:"value_id"`
}

type ResolveCompleteVariantView struct {
	VariantId       string                     `json:"variant_id,omitempty"`
	BaseVariantId   string                     `json:"base_variant_id,omitempty"`
	CombinationKey  string                     `json:"combination_key"`
	Outcome         ResolveOutcome             `json:"outcome"`
	Materialized    bool                       `json:"materialized"`
	NeverSelections []ResolveSelectionPairView `json:"never_selections"`
}

func NewResolveCompleteVariantView(data ResolveCompleteVariantResultData) ResolveCompleteVariantView {
	never := make([]ResolveSelectionPairView, 0, len(data.NeverSelections))
	for _, selection := range data.NeverSelections {
		never = append(never, ResolveSelectionPairView{AttributeId: selection.AttributeId, ValueId: selection.ValueId})
	}
	return ResolveCompleteVariantView{
		VariantId:       data.VariantId,
		BaseVariantId:   data.BaseVariantId,
		CombinationKey:  data.CombinationKey,
		Outcome:         data.Outcome,
		Materialized:    data.Outcome == ResolveOutcomeCreated,
		NeverSelections: never,
	}
}
