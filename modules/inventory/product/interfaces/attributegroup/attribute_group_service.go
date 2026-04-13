package attributegroup

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type AttributeGroupService interface {
	CreateAttributeGroup(ctx corectx.Context, cmd CreateAttributeGroupCommand) (*CreateAttributeGroupResult, error)
	DeleteAttributeGroup(ctx corectx.Context, cmd DeleteAttributeGroupCommand) (*DeleteAttributeGroupResult, error)
	AttributeGroupExists(ctx corectx.Context, query AttributeGroupExistsQuery) (*AttributeGroupExistsResult, error)
	GetAttributeGroup(ctx corectx.Context, query GetAttributeGroupQuery) (*GetAttributeGroupResult, error)
	SearchAttributeGroups(ctx corectx.Context, query SearchAttributeGroupsQuery) (*SearchAttributeGroupsResult, error)
	UpdateAttributeGroup(ctx corectx.Context, cmd UpdateAttributeGroupCommand) (*dyn.OpResult[dyn.MutateResultData], error)
}
