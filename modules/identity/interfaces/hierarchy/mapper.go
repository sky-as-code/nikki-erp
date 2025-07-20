package hierarchy

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (cmd CreateHierarchyLevelCommand) ToHierarchyLevel() *domain.HierarchyLevel {
	hierarchyLevel := &domain.HierarchyLevel{}
	model.MustCopy(cmd, hierarchyLevel)
	return hierarchyLevel
}

func (cmd UpdateHierarchyLevelCommand) ToHierarchyLevel() *domain.HierarchyLevel {
	hierarchyLevel := &domain.HierarchyLevel{}
	model.MustCopy(cmd, hierarchyLevel)
	return hierarchyLevel
}
