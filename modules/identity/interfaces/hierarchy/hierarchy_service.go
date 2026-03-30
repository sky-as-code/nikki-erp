package hierarchy

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type HierarchyService interface {
	CreateHierarchyLevel(ctx corectx.Context, cmd CreateHierarchyLevelCommand) (*CreateHierarchyLevelResult, error)
	DeleteHierarchyLevel(ctx corectx.Context, cmd DeleteHierarchyLevelCommand) (*DeleteHierarchyLevelResult, error)
	GetHierarchyLevel(ctx corectx.Context, query GetHierarchyLevelQuery) (*GetHierarchyLevelResult, error)
	HierarchyLevelExists(ctx corectx.Context, cmd HierarchyLevelExistsQuery) (*HierarchyLevelExistsResult, error)
	ManageHierarchyLevelUsers(ctx corectx.Context, cmd ManageHierarchyLevelUsersCommand) (*ManageHierarchyLevelUsersResult, error)
	SearchHierarchyLevels(ctx corectx.Context, query SearchHierarchyLevelsQuery) (*SearchHierarchyLevelsResult, error)
	UpdateHierarchyLevel(ctx corectx.Context, cmd UpdateHierarchyLevelCommand) (*UpdateHierarchyLevelResult, error)
}
