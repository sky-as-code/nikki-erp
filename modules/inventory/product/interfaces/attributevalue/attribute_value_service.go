package attributevalue

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type AttributeValueService interface {
	CreateAttributeValue(ctx crud.Context, cmd CreateAttributeValueCommand) (*CreateAttributeValueResult, error)
	UpdateAttributeValue(ctx crud.Context, cmd UpdateAttributeValueCommand) (*UpdateAttributeValueResult, error)
	DeleteAttributeValue(ctx crud.Context, cmd DeleteAttributeValueCommand) (*DeleteAttributeValueResult, error)
	GetAttributeValueById(ctx crud.Context, query GetAttributeValueByIdQuery) (*GetAttributeValueByIdResult, error)
	SearchAttributeValues(ctx crud.Context, query SearchAttributeValuesQuery) (*SearchAttributeValuesResult, error)
}
