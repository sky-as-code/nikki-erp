package repository

// func BuildOrganizationDescriptor() *orm.EntityDescriptor {
// 	e := ent.Organization{}
// 	b := orm.DescribeEntity(entOrg.Label).
// 		Aliases("organizations").
// 		Field(entOrg.FieldID, e.ID).
// 		Field(entOrg.FieldCreatedAt, e.CreatedAt).
// 		Field(entOrg.FieldSlug, e.Slug).
// 		Field(entOrg.FieldDisplayName, e.DisplayName).
// 		Field(entOrg.FieldEtag, e.Etag).
// 		Field(entOrg.FieldStatus, e.Status).
// 		Field(entOrg.FieldUpdatedAt, e.UpdatedAt).
// 		Edge(entOrg.EdgeUsers, orm.ToEdgePredicate(entOrg.HasUsersWith)).
// 		Edge(entOrg.EdgeHierarchies, orm.ToEdgePredicate(entOrg.HasHierarchiesWith))
// 	return b.Descriptor()
// }

// func BuildHierarchyLevelDescriptor() *orm.EntityDescriptor {
// 	e := ent.HierarchyLevel{}
// 	b := orm.DescribeEntity(entHierarchy.Label).
// 		Aliases("hierarchy_levels", "hierarchies").
// 		Field(entHierarchy.FieldID, e.ID).
// 		Field(entHierarchy.FieldCreatedAt, e.CreatedAt).
// 		Field(entHierarchy.FieldName, e.Name).
// 		Field(entHierarchy.FieldOrgID, e.OrgID).
// 		Field(entHierarchy.FieldParentID, e.ParentID).
// 		Field(entHierarchy.FieldEtag, e.Etag).
// 		Field(entHierarchy.FieldUpdatedAt, e.UpdatedAt).
// 		Edge(entHierarchy.EdgeUsers, orm.ToEdgePredicate(entHierarchy.HasUsersWith)).
// 		Edge(entHierarchy.EdgeOrg, orm.ToEdgePredicate(entHierarchy.HasOrgWith)).
// 		Edge(entHierarchy.EdgeChildren, orm.ToEdgePredicate(entHierarchy.HasChildrenWith))
// 	return b.Descriptor()
// }
