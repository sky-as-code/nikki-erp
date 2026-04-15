package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	PurchaseRequestSchemaName = "purchase.purchase_request"
	PurchaseRequestItemSchemaName = "purchase.purchase_request_item"

	PurchaseRequestFieldId                    = basemodel.FieldId
	PurchaseRequestFieldCode                  = "code"
	PurchaseRequestFieldSummary               = "summary"
	PurchaseRequestFieldRequestedAt           = "requested_at"
	PurchaseRequestFieldRequestedBy           = "requested_by"
	PurchaseRequestFieldRequestingHierarchyId = "requesting_hierarchy_id"
	PurchaseRequestFieldRequiredAt            = "required_at"
	PurchaseRequestFieldJustification         = "justification"
	PurchaseRequestFieldNote                  = "note"
	PurchaseRequestFieldStatus                = "status"
	PurchaseRequestFieldApprovalLevel         = "approval_level"
	PurchaseRequestFieldPriority              = "priority"
	PurchaseRequestFieldConversionType        = "conversion_type"

	PurchaseRequestStatusDraft            = "draft"
	PurchaseRequestStatusPendingApproval  = "pending_approval"
	PurchaseRequestStatusApproved         = "approved"
	PurchaseRequestStatusRejected         = "rejected"
	PurchaseRequestStatusCancelled        = "cancelled"
	PurchaseRequestStatusConvertedToRfq   = "converted_to_rfq"
	PurchaseRequestStatusConvertedToPo    = "converted_to_po"
	PurchaseRequestPriorityNormal         = "normal"
	PurchaseRequestPriorityUrgent         = "urgent"

	PurchaseRequestEdgeItems = "items"

	PurchaseRequestItemFieldId                = basemodel.FieldId
	PurchaseRequestItemFieldPurchaseRequestId = "purchase_request_id"
	PurchaseRequestItemFieldProductSku        = "product_sku"
	PurchaseRequestItemFieldUnit              = "unit"
	PurchaseRequestItemFieldQuantity          = "quantity"
	PurchaseRequestItemFieldPurpose           = "purpose"
	PurchaseRequestItemFieldNote              = "note"
)

func PurchaseRequestSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PurchaseRequestSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Purchase request"}).
		TableName("purchase_requests").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldCode).DataType(dmodel.FieldDataTypeString(1, 50)).RequiredForCreate().Unique()).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldSummary).DataType(dmodel.FieldDataTypeLangJson(0, 100)).RequiredForCreate()).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldRequestedAt).DataType(dmodel.FieldDataTypeDateTime()).RequiredForCreate()).
		Field(basemodel.DefineFieldId(PurchaseRequestFieldRequestedBy).RequiredForCreate()).
		Field(basemodel.DefineFieldId(PurchaseRequestFieldRequestingHierarchyId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldRequiredAt).DataType(dmodel.FieldDataTypeDateTime()).RequiredForCreate()).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldJustification).DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH))).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldNote).DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH))).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldStatus).DataType(dmodel.FieldDataTypeEnumString([]string{
			PurchaseRequestStatusDraft,
			PurchaseRequestStatusPendingApproval,
			PurchaseRequestStatusApproved,
			PurchaseRequestStatusRejected,
			PurchaseRequestStatusCancelled,
			PurchaseRequestStatusConvertedToRfq,
			PurchaseRequestStatusConvertedToPo,
		})).Default(PurchaseRequestStatusDraft).RequiredForCreate()).
		Field(
			dmodel.DefineField().
				Name(PurchaseRequestFieldApprovalLevel).
				DataType(dmodel.FieldDataTypeInt32(0, 100)).
				Default(0),
		).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldPriority).DataType(
			dmodel.FieldDataTypeEnumString([]string{PurchaseRequestPriorityNormal, PurchaseRequestPriorityUrgent}),
		).Default(PurchaseRequestPriorityNormal)).
		Field(dmodel.DefineField().Name(PurchaseRequestFieldConversionType).DataType(
			dmodel.FieldDataTypeEnumString([]string{"rfq", "po"}),
		)).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(PurchaseRequestEdgeItems).
				OneToMany(PurchaseRequestItemSchemaName, dmodel.DynamicFields{
					PurchaseRequestItemFieldPurchaseRequestId: PurchaseRequestFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

func PurchaseRequestItemSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PurchaseRequestItemSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Purchase request item"}).
		TableName("purchase_request_items").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(PurchaseRequestItemFieldPurchaseRequestId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(PurchaseRequestItemFieldProductSku).DataType(dmodel.FieldDataTypeString(1, 120)).RequiredForCreate()).
		Field(dmodel.DefineField().Name(PurchaseRequestItemFieldUnit).DataType(dmodel.FieldDataTypeString(1, 80)).RequiredForCreate()).
		Field(
			dmodel.DefineField().
				Name(PurchaseRequestItemFieldQuantity).
				DataType(dmodel.FieldDataTypeDecimal("0.0001", "99999999999999.9999", 4)).
				RequiredForCreate(),
		).
		Field(dmodel.DefineField().Name(PurchaseRequestItemFieldPurpose).DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH))).
		Field(dmodel.DefineField().Name(PurchaseRequestItemFieldNote).DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH))).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type PurchaseRequest struct{ basemodel.DynamicModelBase }
type PurchaseRequestItem struct{ basemodel.DynamicModelBase }

func NewPurchaseRequest() *PurchaseRequest { return &PurchaseRequest{basemodel.NewDynamicModel()} }
func NewPurchaseRequestFrom(src dmodel.DynamicFields) *PurchaseRequest {
	return &PurchaseRequest{basemodel.NewDynamicModel(src)}
}
