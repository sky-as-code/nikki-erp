package attribute

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type AttributeService interface {
	CreateAttribute(ctx corectx.Context, cmd CreateAttributeCommand) (*CreateAttributeResult, error)
	DeleteAttribute(ctx corectx.Context, cmd DeleteAttributeCommand) (*DeleteAttributeResult, error)
	AttributeExists(ctx corectx.Context, query AttributeExistsQuery) (*AttributeExistsResult, error)
	GetAttribute(ctx corectx.Context, query GetAttributeQuery) (*GetAttributeResult, error)
	SearchAttributes(ctx corectx.Context, query SearchAttributesQuery) (*SearchAttributesResult, error)
	UpdateAttribute(ctx corectx.Context, cmd UpdateAttributeCommand) (*dyn.OpResult[dyn.MutateResultData], error)
}
