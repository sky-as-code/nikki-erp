package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type AttributeRepository interface {
	Create(ctx crud.Context, attribute *Attribute) (*Attribute, error)
	Update(ctx crud.Context, attribute *Attribute, prevEtag model.Etag) (*Attribute, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*Attribute, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[Attribute], error)
}

type AttributeService interface {
	CreateAttribute(ctx crud.Context, cmd CreateAttributeCommand) (*CreateAttributeResult, error)
	UpdateAttribute(ctx crud.Context, cmd UpdateAttributeCommand) (*UpdateAttributeResult, error)
	DeleteAttribute(ctx crud.Context, cmd DeleteAttributeCommand) (*DeleteAttributeResult, error)
	GetAttributeById(ctx crud.Context, query GetAttributeByIdQuery) (*GetAttributeByIdResult, error)
	SearchAttributes(ctx crud.Context, query SearchAttributesQuery) (*SearchAttributesResult, error)
}

type DeleteParam = DeleteAttributeCommand
type FindByIdParam = GetAttributeByIdQuery
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
