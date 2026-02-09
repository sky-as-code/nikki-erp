package attributevalue

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type AttributeValueRepository interface {
	Create(ctx crud.Context, attributeValue *domain.AttributeValue) (*domain.AttributeValue, error)
	CreateAndLinkVariant(ctx crud.Context, attributeValue *domain.AttributeValue, variantId model.Id) (*domain.AttributeValue, error)
	Update(ctx crud.Context, attributeValue *domain.AttributeValue, prevEtag model.Etag) (*domain.AttributeValue, error)
	LinkVariantToExisting(ctx crud.Context, attributeValueId model.Id, variantId model.Id, prevEtag model.Etag) (*domain.AttributeValue, bool, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*domain.AttributeValue, error)
	FindByValueRef(ctx crud.Context, attributeValue *domain.AttributeValue, dataType string) (*domain.AttributeValue, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.AttributeValue], error)
}

type DeleteParam = DeleteAttributeValueCommand
type FindByIdParam = GetAttributeValueByIdQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
