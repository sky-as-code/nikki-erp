package hierarchy

import (
	"context"
)

type HierarchyService interface {
	AddRemoveUsers(ctx context.Context, cmd AddRemoveUsersCommand) (*AddRemoveUsersResult, error)
	CreateHierarchyLevel(ctx context.Context, cmd CreateHierarchyLevelCommand) (*CreateHierarchyLevelResult, error)
	DeleteHierarchyLevel(ctx context.Context, cmd DeleteHierarchyLevelCommand) (*DeleteHierarchyLevelResult, error)
	GetHierarchyLevelById(ctx context.Context, query GetHierarchyLevelByIdQuery) (*GetHierarchyLevelByIdResult, error)
	SearchHierarchyLevels(ctx context.Context, query SearchHierarchyLevelsQuery) (*SearchHierarchyLevelsResult, error)
	UpdateHierarchyLevel(ctx context.Context, cmd UpdateHierarchyLevelCommand) (*UpdateHierarchyLevelResult, error)
}
