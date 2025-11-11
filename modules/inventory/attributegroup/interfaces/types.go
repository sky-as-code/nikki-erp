package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type AttributeGroupRepository interface {
	Create(ctx crud.Context, attributeGroup *AttributeGroup) (*AttributeGroup, error)
	Update(ctx crud.Context, attributeGroup *AttributeGroup, prevEtag model.Etag) (*AttributeGroup, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*AttributeGroup, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[AttributeGroup], error)
}

type AttributeGroupService interface {
	CreateAttributeGroup(ctx crud.Context, cmd CreateAttributeGroupCommand) (*CreateAttributeGroupResult, error)
	UpdateAttributeGroup(ctx crud.Context, cmd UpdateAttributeGroupCommand) (*UpdateAttributeGroupResult, error)
	DeleteAttributeGroup(ctx crud.Context, cmd DeleteAttributeGroupCommand) (*DeleteAttributeGroupResult, error)
	GetAttributeGroupById(ctx crud.Context, query GetAttributeGroupByIdQuery) (*GetAttributeGroupByIdResult, error)
	SearchAttributeGroups(ctx crud.Context, query SearchAttributeGroupsQuery) (*SearchAttributeGroupsResult, error)
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
