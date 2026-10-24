package product

import (
	"encoding/json"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The attribute configuration of a template rides the template's own create and update, and the
// preview action, as the FULL desired state in catalogue ids. Neither field is a template column:
// the domain service lifts them before schema validation. Omitting a field leaves that part of
// the configuration unchanged.
const (
	ProductTemplateFieldAttributeLines = "attribute_lines"
	ProductTemplateFieldExclusions     = "exclusions"

	// ProductTemplateFieldVariantDeleteAllowed is set by the application service only, after a
	// non-failing check of the caller's product_variant delete permission: an invalid variant
	// that nothing uses is deleted when true and archived otherwise. Any client value is
	// overwritten before the domain service reads it.
	ProductTemplateFieldVariantDeleteAllowed = "variant_delete_allowed"

	// ProductTemplateFieldAttributeConfiguration is the computed read of the same configuration.
	ProductTemplateFieldAttributeConfiguration = "attribute_configuration"
)

// ProductTemplateConfigurationMaxLines caps the lines of one payload.
const ProductTemplateConfigurationMaxLines = 100

// ProductTemplateConfigurationMaxExclusions caps the exclusions of one payload.
const ProductTemplateConfigurationMaxExclusions = 1000

// AttributeLineInput is one desired line: a catalogue attribute and the catalogue values the
// template selects from it, at least one.
type AttributeLineInput struct {
	AttributeId string   `json:"attribute_id"`
	ValueIds    []string `json:"value_ids"`
	Sequence    *int32   `json:"sequence,omitempty"`
}

// ExclusionInput forbids two catalogue values from meeting in one variant. Unordered.
type ExclusionInput struct {
	LeftValueId  string `json:"left_value_id"`
	RightValueId string `json:"right_value_id"`
}

// TemplateConfigurationInput is the lifted payload. A nil slice pointer means "unchanged".
type TemplateConfigurationInput struct {
	AttributeLines *[]AttributeLineInput
	Exclusions     *[]ExclusionInput
}

func (this TemplateConfigurationInput) IsEmpty() bool {
	return this.AttributeLines == nil && this.Exclusions == nil
}

// LiftTemplateConfiguration removes `attribute_lines`, `exclusions` and the delete flag from a
// template command, so the template schema never sees them. input is nil when neither
// configuration field was sent; malformed names a field that could not be decoded.
func LiftTemplateConfiguration(
	fields dmodel.DynamicFields,
) (input *TemplateConfigurationInput, deleteAllowed bool, malformed string) {
	deleteAllowed, _ = fields[ProductTemplateFieldVariantDeleteAllowed].(bool)
	delete(fields, ProductTemplateFieldVariantDeleteAllowed)
	delete(fields, ProductTemplateFieldAttributeConfiguration)

	lifted := TemplateConfigurationInput{}
	if !liftJsonField(fields, ProductTemplateFieldAttributeLines, &lifted.AttributeLines) {
		malformed = ProductTemplateFieldAttributeLines
	}
	if !liftJsonField(fields, ProductTemplateFieldExclusions, &lifted.Exclusions) {
		malformed = ProductTemplateFieldExclusions
	}
	if lifted.IsEmpty() {
		return nil, deleteAllowed, malformed
	}
	return &lifted, deleteAllowed, malformed
}

// liftJsonField decodes and removes one nested field. An absent or null field leaves target nil
// and reports success.
func liftJsonField[T any](fields dmodel.DynamicFields, field string, target **[]T) bool {
	raw, found := fields[field]
	if !found {
		return true
	}
	delete(fields, field)
	if raw == nil {
		return true
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return false
	}
	decoded := []T{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return false
	}
	*target = &decoded
	return true
}

// The preview action: `POST product_template/:id/preview_attribute_configuration` with
// {org_id, etag, attribute_lines, exclusions}.
type (
	PreviewAttributeConfigurationCommand = dmodel.DynamicFields
	PreviewAttributeConfigurationResult  = dyn.OpResult[any]
)

// PreviewTemplateConfigurationQuery is the domain dry run. Etag must match the template's.
type PreviewTemplateConfigurationQuery struct {
	TemplateId    string
	Etag          string
	Input         TemplateConfigurationInput
	DeleteAllowed bool
}

// ProductTemplateConfigurationService is the configuration surface of the template domain
// service beside ProductService.
type ProductTemplateConfigurationService interface {
	PreviewTemplateConfiguration(
		ctx corectx.Context, query PreviewTemplateConfigurationQuery,
	) (*dyn.OpResult[TemplateConfigurationPreview], error)
}

// TemplateConfigurationPreview is the dry-run answer: what applying the payload would do to every
// variant record, which is re-decided at apply time (usage may change in between).
type TemplateConfigurationPreview struct {
	Etag   string        `json:"etag"`
	Counts PreviewCounts `json:"counts"`

	PreCreateKind         string           `json:"pre_create_kind,omitempty"`
	PreCreateCount        int              `json:"pre_create_count"`
	PossibleCompleteCount *int             `json:"possible_complete_count"`
	PossibleOverThreshold bool             `json:"possible_over_threshold"`
	InactiveExclusions    []ExclusionInput `json:"inactive_exclusions"`

	Keep      []PreviewVariant `json:"keep"`
	Extend    []PreviewVariant `json:"extend"`
	Restore   []PreviewVariant `json:"restore"`
	Create    []PreviewVariant `json:"create"`
	Archive   []PreviewVariant `json:"archive"`
	Delete    []PreviewVariant `json:"delete"`
	Untouched []PreviewVariant `json:"untouched"`

	// ChangesIdentity is true when anything but keep/create would happen, so a client knows a
	// confirmation is due.
	ChangesIdentity bool `json:"changes_identity"`
}

type PreviewCounts struct {
	Keep      int `json:"keep"`
	Extend    int `json:"extend"`
	Restore   int `json:"restore"`
	Create    int `json:"create"`
	Archive   int `json:"archive"`
	Delete    int `json:"delete"`
	Untouched int `json:"untouched"`
}

// PreviewVariant is one record of the preview. VariantId is empty for a create. Stock figures and
// SKU/barcode are filled for archive and delete, the outcomes that take a record out of use.
//
// DisplayName reads as the variant's display_name, "Coca Cola (No sugar, Big)": an existing
// record's current name, or the name a created one will have. NewDisplayName is set on an extend
// only, naming the record after its key grows.
type PreviewVariant struct {
	VariantId          string   `json:"variant_id,omitempty"`
	DisplayName        string   `json:"display_name"`
	NewDisplayName     string   `json:"new_display_name,omitempty"`
	VariantKind        string   `json:"variant_kind"`
	CombinationKey     string   `json:"combination_key"`
	NewCombinationKey  string   `json:"new_combination_key,omitempty"`
	BaseCombinationKey string   `json:"base_combination_key,omitempty"`
	ValueIds           []string `json:"value_ids"`
	IsArchived         bool     `json:"is_archived"`
	ArchiveSource      string   `json:"archive_source,omitempty"`
	Sku                *string  `json:"sku,omitempty"`
	PrimaryBarcode     *string  `json:"primary_barcode,omitempty"`
	OnHandQuantity     string   `json:"on_hand_quantity,omitempty"`
	ReservedQuantity   string   `json:"reserved_quantity,omitempty"`
	// ArchiveReason says why an invalid record is archived rather than deleted:
	// "ever_used", "used_by_modules" or "no_delete_permission".
	ArchiveReason string   `json:"archive_reason,omitempty"`
	UsedBy        []string `json:"used_by,omitempty"`
}
