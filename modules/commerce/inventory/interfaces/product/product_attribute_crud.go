package product

import (
	"encoding/json"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductAttributeRepository reads and writes product_attribute rows.
type ProductAttributeRepository interface {
	composable.CrudRepository
}

// ProductAttributeDomainService is the CRUD of the resource plus its own rules.
type ProductAttributeDomainService interface {
	composable.CrudDomainService
}

// ProductAttributeApplicationService is the authorized surface the REST handler serves.
type ProductAttributeApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductAttributeCommand      = composable.CreateCommand
	UpdateProductAttributeCommand      = composable.UpdateCommand
	DeleteProductAttributeCommand      = composable.DeleteCommand
	SetProductAttributeArchivedCommand = composable.SetArchivedCommand
	GetProductAttributeByIdQuery       = composable.GetByIdQuery
	SearchProductAttributesQuery       = composable.SearchQuery
	ProductAttributeExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductAttributeResult      = composable.CreateResult
	UpdateProductAttributeResult      = composable.MutateResult
	DeleteProductAttributeResult      = composable.MutateResult
	SetProductAttributeArchivedResult = composable.MutateResult
	GetProductAttributeByIdResult     = composable.GetOneResult
	SearchProductAttributesResult     = composable.SearchResult
	ProductAttributeExistsResult      = composable.ExistsResult
)

// ProductAttributeFieldValues carries the inline values editor on an attribute create or update.
// It is not an attribute column: the domain service lifts it before schema validation.
const ProductAttributeFieldValues = "values"

// ProductAttributeValuesDeltaMaxItems caps the rows of one delta: every row is a write and a
// sibling-name check inside the attribute's transaction.
const ProductAttributeValuesDeltaMaxItems = 500

// ProductAttributeValuesDelta rides an attribute create or update as
// `values: {add: [...], update: [...], remove: [ids]}`. Add rows carry the value's own fields
// (name, code, sequence, color_hex); update rows add `id` and `etag`. The parent's
// attribute_id and org_id are always the server's to set. A create accepts `add` only.
type ProductAttributeValuesDelta struct {
	Add    []dmodel.DynamicFields `json:"add"`
	Update []dmodel.DynamicFields `json:"update"`
	Remove []model.Id             `json:"remove"`
}

func (this ProductAttributeValuesDelta) IsEmpty() bool {
	return len(this.Add) == 0 && len(this.Update) == 0 && len(this.Remove) == 0
}

// LiftProductAttributeValuesDelta removes `values` from an attribute command, so the attribute
// schema never sees it, and returns the decoded delta. malformed reports a value that is not a
// delta.
func LiftProductAttributeValuesDelta(
	fields dmodel.DynamicFields,
) (delta *ProductAttributeValuesDelta, malformed bool) {
	raw, found := fields[ProductAttributeFieldValues]
	if !found {
		return nil, false
	}
	delete(fields, ProductAttributeFieldValues)
	if raw == nil {
		return nil, false
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, true
	}
	delta = &ProductAttributeValuesDelta{}
	if err := json.Unmarshal(encoded, delta); err != nil {
		return nil, true
	}
	return delta, false
}
