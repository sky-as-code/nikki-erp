package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type AttributeValueRepository interface {
	Create(ctx crud.Context, attributeValue *AttributeValue) (*AttributeValue, error)
	Update(ctx crud.Context, attributeValue *AttributeValue, prevEtag model.Etag) (*AttributeValue, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*AttributeValue, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[AttributeValue], error)
}

type AttributeValueService interface {
	CreateAttributeValue(ctx crud.Context, cmd CreateAttributeValueCommand) (*CreateAttributeValueResult, error)
	UpdateAttributeValue(ctx crud.Context, cmd UpdateAttributeValueCommand) (*UpdateAttributeValueResult, error)
	DeleteAttributeValue(ctx crud.Context, cmd DeleteAttributeValueCommand) (*DeleteAttributeValueResult, error)
	GetAttributeValueById(ctx crud.Context, query GetAttributeValueByIdQuery) (*GetAttributeValueByIdResult, error)
	SearchAttributeValues(ctx crud.Context, query SearchAttributeValuesQuery) (*SearchAttributeValuesResult, error)
}

type DeleteParam = DeleteAttributeValueCommand
type FindByIdParam = GetAttributeValueByIdQuery
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
