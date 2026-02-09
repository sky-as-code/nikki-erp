package attribute

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type AttributeRepository interface {
	Create(ctx crud.Context, attribute *domain.Attribute) (*domain.Attribute, error)
	Update(ctx crud.Context, attribute *domain.Attribute, prevEtag model.Etag) (*domain.Attribute, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*domain.Attribute, error)
	FindByCodeName(ctx crud.Context, query FindByCodeNameParam) (*domain.Attribute, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Attribute], error)
}

type DeleteParam = DeleteAttributeCommand
type FindByIdParam = GetAttributeByIdQuery
type FindByCodeNameParam = GetAttributeByCodeName

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
