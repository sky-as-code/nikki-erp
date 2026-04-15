package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	VendorSchemaName = "purchase.vendor"

	VendorFieldId           = basemodel.FieldId
	VendorFieldStatus       = "status"
	VendorFieldStatusReason = "status_reason"

	VendorStatusProposed    = "proposed"
	VendorStatusActive      = "active"
	VendorStatusSuspended   = "suspended"
	VendorStatusBlacklisted = "blacklisted"
)

func VendorSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(VendorSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Vendor"}).
		TableName("purchase_vendors").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(VendorFieldStatus).DataType(
			dmodel.FieldDataTypeEnumString([]string{
				VendorStatusProposed, VendorStatusActive, VendorStatusSuspended, VendorStatusBlacklisted,
			}),
		).Default(VendorStatusProposed).RequiredForCreate()).
		Field(dmodel.DefineField().Name(VendorFieldStatusReason).DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH))).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type Vendor struct{ basemodel.DynamicModelBase }

func NewVendor() *Vendor { return &Vendor{basemodel.NewDynamicModel()} }
