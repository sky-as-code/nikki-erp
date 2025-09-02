package hierarchy

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type HierarchyService interface {
	AddRemoveUsers(ctx crud.Context, cmd AddRemoveUsersCommand) (*AddRemoveUsersResult, error)
	CreateHierarchyLevel(ctx crud.Context, cmd CreateHierarchyLevelCommand) (*CreateHierarchyLevelResult, error)
	DeleteHierarchyLevel(ctx crud.Context, cmd DeleteHierarchyLevelCommand) (*DeleteHierarchyLevelResult, error)
	GetHierarchyLevelById(ctx crud.Context, query GetHierarchyLevelByIdQuery) (*GetHierarchyLevelByIdResult, error)
	SearchHierarchyLevels(ctx crud.Context, query SearchHierarchyLevelsQuery) (*SearchHierarchyLevelsResult, error)
	UpdateHierarchyLevel(ctx crud.Context, cmd UpdateHierarchyLevelCommand) (*UpdateHierarchyLevelResult, error)
}
