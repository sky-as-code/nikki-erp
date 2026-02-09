package attributegroup

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type AttributeGroupService interface {
	CreateAttributeGroup(ctx crud.Context, cmd CreateAttributeGroupCommand) (*CreateAttributeGroupResult, error)
	UpdateAttributeGroup(ctx crud.Context, cmd UpdateAttributeGroupCommand) (*UpdateAttributeGroupResult, error)
	DeleteAttributeGroup(ctx crud.Context, cmd DeleteAttributeGroupCommand) (*DeleteAttributeGroupResult, error)
	GetAttributeGroupById(ctx crud.Context, query GetAttributeGroupByIdQuery) (*GetAttributeGroupByIdResult, error)
	SearchAttributeGroups(ctx crud.Context, query SearchAttributeGroupsQuery) (*SearchAttributeGroupsResult, error)
}
