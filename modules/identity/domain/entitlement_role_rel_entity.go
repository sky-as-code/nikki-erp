package domain

// dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

// const (
// 	EntitlementRoleRelSchemaName = "authorize.entitlement_role_rel"
// 	EntRoleRelFieldEntitlementId = "entitlement_id"
// 	EntRoleRelFieldRoleId        = "role_id"
// )

// func EntitlementRoleRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
// 	return dmodel.DefineModel(EntitlementRoleRelSchemaName).
// 		TableName("authz_entitlement_role_rel").
// 		ShouldBuildDb().
// 		Field(
// 			dmodel.DefineField().Name(EntRoleRelFieldEntitlementId).
// 				DataType(dmodel.FieldDataTypeUlid()).
// 				RequiredForCreate().
// 				PrimaryKey(),
// 		).
// 		Field(
// 			dmodel.DefineField().Name(EntRoleRelFieldRoleId).
// 				DataType(dmodel.FieldDataTypeUlid()).
// 				RequiredForCreate().
// 				PrimaryKey(),
// 		)
// }
