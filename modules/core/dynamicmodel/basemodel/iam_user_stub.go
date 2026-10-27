package basemodel

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

const iamUserSchemaName = "iam_user"

// RegisterIamUserStub registers a minimal iam_user, for a test that finalizes a module's schemas
// without loading IAM: every traceable mixin carries an edge to it, and finalizing fails on an edge
// whose destination is missing. It does nothing when iam_user is already registered.
func RegisterIamUserStub() error {
	if dmodel.GetSchemaRegistry().Get(iamUserSchemaName) != nil {
		return nil
	}
	return dmodel.RegisterSchemaB(dmodel.DefineModel(iamUserSchemaName).
		TableName("iam_users").
		ShouldBuildDb().
		Field(DefineFieldId(FieldId).PrimaryKey()).
		Field(dmodel.DefineField().Name("display_name").DataType(dmodel.FieldDataTypeString(1, 200))).
		RecordLabelField("display_name"))
}
