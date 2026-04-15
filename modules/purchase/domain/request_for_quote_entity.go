package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RequestForQuoteSchemaName = "purchase.request_for_quote"
	RequestForQuoteFieldId     = basemodel.FieldId
	RequestForQuoteFieldCode   = "code"
	RequestForQuoteFieldStatus = "status"
)

func RequestForQuoteSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RequestForQuoteSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Request for quote"}).
		TableName("purchase_request_for_quotes").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(RequestForQuoteFieldCode).DataType(dmodel.FieldDataTypeString(1, 50)).RequiredForCreate().Unique()).
		Field(dmodel.DefineField().Name(RequestForQuoteFieldStatus).DataType(dmodel.FieldDataTypeEnumString([]string{
			"draft", "published", "closed",
		})).Default("draft").RequiredForCreate()).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type RequestForQuote struct{ basemodel.DynamicModelBase }

func NewRequestForQuote() *RequestForQuote { return &RequestForQuote{basemodel.NewDynamicModel()} }
func NewRequestForQuoteFrom(src dmodel.DynamicFields) *RequestForQuote {
	return &RequestForQuote{basemodel.NewDynamicModel(src)}
}
