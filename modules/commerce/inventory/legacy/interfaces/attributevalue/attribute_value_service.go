package attributevalue

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type AttributeValueService interface {
	CreateAttributeValue(ctx corectx.Context, cmd CreateAttributeValueCommand) (*CreateAttributeValueResult, error)
	DeleteAttributeValue(ctx corectx.Context, cmd DeleteAttributeValueCommand) (*DeleteAttributeValueResult, error)
	AttributeValueExists(ctx corectx.Context, query AttributeValueExistsQuery) (*AttributeValueExistsResult, error)
	GetAttributeValue(ctx corectx.Context, query GetAttributeValueQuery) (*GetAttributeValueResult, error)
	SearchAttributeValues(ctx corectx.Context, query SearchAttributeValuesQuery) (*SearchAttributeValuesResult, error)
	UpdateAttributeValue(ctx corectx.Context, cmd UpdateAttributeValueCommand) (*dyn.OpResult[dyn.MutateResultData], error)
	GetAttributeValueIdsByVariantId(ctx corectx.Context, variantId model.Id) ([]model.Id, error)
	FindOrCreateAttributeValue(ctx corectx.Context, attribute *domain.Attribute, value any, codeName string, vErrs *ft.ClientErrors) (*model.Id, error)
}
