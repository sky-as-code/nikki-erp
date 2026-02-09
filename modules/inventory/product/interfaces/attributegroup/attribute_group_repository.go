package attributegroup

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type AttributeGroupRepository interface {
	Create(ctx crud.Context, attributeGroup *domain.AttributeGroup) (*domain.AttributeGroup, error)
	Update(ctx crud.Context, attributeGroup *domain.AttributeGroup, prevEtag model.Etag) (*domain.AttributeGroup, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*domain.AttributeGroup, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.AttributeGroup], error)
}

type DeleteParam = DeleteAttributeGroupCommand
type FindByIdParam = GetAttributeGroupByIdQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
	ProductId *string
}
