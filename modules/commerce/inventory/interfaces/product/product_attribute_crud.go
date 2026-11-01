package product

import (
	"encoding/json"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductAttributeRepository reads and writes product_attribute rows.
type ProductAttributeRepository interface {
	dynamicresource.CrudRepository
}

// ProductAttributeDomainService is the CRUD of the resource plus its own rules.
type ProductAttributeDomainService interface {
	dynamicresource.CrudDomainService
}

// ProductAttributeApplicationService is the authorized surface the REST handler serves.
type ProductAttributeApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductAttributeCommand      = dynamicresource.CreateCommand
	UpdateProductAttributeCommand      = dynamicresource.UpdateCommand
	DeleteProductAttributeCommand      = dynamicresource.DeleteCommand
	SetProductAttributeArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductAttributeByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductAttributesQuery       = dynamicresource.SearchQuery
	ProductAttributeExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductAttributeResult      = dynamicresource.CreateResult
	UpdateProductAttributeResult      = dynamicresource.MutateResult
	DeleteProductAttributeResult      = dynamicresource.MutateResult
	SetProductAttributeArchivedResult = dynamicresource.MutateResult
	GetProductAttributeByIdResult     = dynamicresource.GetOneResult
	SearchProductAttributesResult     = dynamicresource.SearchResult
	ProductAttributeExistsResult      = dynamicresource.ExistsResult
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
