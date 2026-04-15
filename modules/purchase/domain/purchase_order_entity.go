package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	PurchaseOrderSchemaName     = "purchase.purchase_order"
	PurchaseOrderItemSchemaName = "purchase.purchase_order_item"

	PurchaseOrderFieldId      = basemodel.FieldId
	PurchaseOrderFieldCode    = "code"
	PurchaseOrderFieldStatus  = "status"
	PurchaseOrderEdgeItems    = "items"
	PurchaseOrderItemFieldId  = basemodel.FieldId
	PurchaseOrderItemFieldPoId = "purchase_order_id"
	PurchaseOrderItemFieldProductId = "product_id"
	PurchaseOrderItemFieldUnit = "unit"
	PurchaseOrderItemFieldQuantity = "quantity"
	PurchaseOrderItemEdgeProduct = "product"
)

func PurchaseOrderSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PurchaseOrderSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Purchase order"}).
		TableName("purchase_orders").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(PurchaseOrderFieldCode).DataType(dmodel.FieldDataTypeString(1, 50)).RequiredForCreate().Unique()).
		Field(dmodel.DefineField().Name(PurchaseOrderFieldStatus).DataType(dmodel.FieldDataTypeEnumString([]string{
			"draft", "approved", "issued", "cancelled",
		})).Default("draft").RequiredForCreate()).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(dmodel.Edge(PurchaseOrderEdgeItems).
			OneToMany(PurchaseOrderItemSchemaName, dmodel.DynamicFields{PurchaseOrderItemFieldPoId: PurchaseOrderFieldId}).
			OnDelete(dmodel.RelationCascadeCascade))
}

func PurchaseOrderItemSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PurchaseOrderItemSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Purchase order item"}).
		TableName("purchase_order_items").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(PurchaseOrderItemFieldPoId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(PurchaseOrderItemFieldProductId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(PurchaseOrderItemFieldUnit).DataType(dmodel.FieldDataTypeString(1, 80)).RequiredForCreate()).
		Field(
			dmodel.DefineField().
				Name(PurchaseOrderItemFieldQuantity).
				DataType(dmodel.FieldDataTypeDecimal("0.0001", "99999999999999.9999", 4)).
				RequiredForCreate(),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(dmodel.Edge(PurchaseOrderItemEdgeProduct).
			ManyToOne("inventory.product", dmodel.DynamicFields{PurchaseOrderItemFieldProductId: basemodel.FieldId}).
			OnDelete(dmodel.RelationCascadeNoAction))
}

type PurchaseOrder struct{ basemodel.DynamicModelBase }
type PurchaseOrderItem struct{ basemodel.DynamicModelBase }

func NewPurchaseOrder() *PurchaseOrder { return &PurchaseOrder{basemodel.NewDynamicModel()} }
func NewPurchaseOrderFrom(src dmodel.DynamicFields) *PurchaseOrder {
	return &PurchaseOrder{basemodel.NewDynamicModel(src)}
}
