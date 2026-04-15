package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RequestForProposalSchemaName = "purchase.request_for_proposal"
	RequestForProposalFieldId     = basemodel.FieldId
	RequestForProposalFieldCode   = "code"
	RequestForProposalFieldStatus = "status"
)

func RequestForProposalSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RequestForProposalSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Request for proposal"}).
		TableName("purchase_request_for_proposals").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(RequestForProposalFieldCode).DataType(dmodel.FieldDataTypeString(1, 50)).RequiredForCreate().Unique()).
		Field(dmodel.DefineField().Name(RequestForProposalFieldStatus).DataType(dmodel.FieldDataTypeEnumString([]string{
			"draft", "published", "closed",
		})).Default("draft").RequiredForCreate()).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type RequestForProposal struct{ basemodel.DynamicModelBase }

func NewRequestForProposal() *RequestForProposal { return &RequestForProposal{basemodel.NewDynamicModel()} }
func NewRequestForProposalFrom(src dmodel.DynamicFields) *RequestForProposal {
	return &RequestForProposal{basemodel.NewDynamicModel(src)}
}
