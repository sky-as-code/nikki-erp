package attribute

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type AttributeService interface {
	CreateAttribute(ctx crud.Context, cmd CreateAttributeCommand) (*CreateAttributeResult, error)
	UpdateAttribute(ctx crud.Context, cmd UpdateAttributeCommand) (*UpdateAttributeResult, error)
	DeleteAttribute(ctx crud.Context, cmd DeleteAttributeCommand) (*DeleteAttributeResult, error)
	GetAttributeById(ctx crud.Context, query GetAttributeByIdQuery) (*GetAttributeByIdResult, error)
	GetAttributeByCodeName(ctx crud.Context, query GetAttributeByCodeName) (*GetAttributeByIdResult, error)
	SearchAttributes(ctx crud.Context, query SearchAttributesQuery) (*SearchAttributesResult, error)
}
