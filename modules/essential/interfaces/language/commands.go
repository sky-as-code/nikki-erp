package language

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateLanguageCommand)(nil)
	req = (*DeleteLanguageCommand)(nil)
	req = (*GetLanguageQuery)(nil)
	req = (*SearchLanguagesQuery)(nil)
	req = (*UpdateLanguageCommand)(nil)
	req = (*LanguageExistsQuery)(nil)
	util.Unused(req)
}

var createLanguageCommandType = cqrs.RequestType{Module: "essential", Submodule: "language", Action: "create"}

type CreateLanguageCommand struct{ models.Language }

func (CreateLanguageCommand) CqrsRequestType() cqrs.RequestType { return createLanguageCommandType }
func (CreateLanguageCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.LanguageSchemaName)
}

type CreateLanguageResult = dyn.OpResult[models.Language]

var updateLanguageCommandType = cqrs.RequestType{Module: "essential", Submodule: "language", Action: "update"}

type UpdateLanguageCommand struct{ models.Language }

func (UpdateLanguageCommand) CqrsRequestType() cqrs.RequestType { return updateLanguageCommandType }
func (UpdateLanguageCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.LanguageSchemaName)
}

type UpdateLanguageResult = dyn.OpResult[dyn.MutateResultData]

var deleteLanguageCommandType = cqrs.RequestType{Module: "essential", Submodule: "language", Action: "delete"}

type DeleteLanguageCommand dyn.DeleteOneCommand

func (DeleteLanguageCommand) CqrsRequestType() cqrs.RequestType { return deleteLanguageCommandType }

type DeleteLanguageResult = dyn.OpResult[dyn.MutateResultData]

var getLanguageQueryType = cqrs.RequestType{Module: "essential", Submodule: "language", Action: "get"}

type GetLanguageQuery dyn.GetOneQuery

func (GetLanguageQuery) CqrsRequestType() cqrs.RequestType { return getLanguageQueryType }

type GetLanguageResult = dyn.OpResult[models.Language]

var searchLanguagesQueryType = cqrs.RequestType{Module: "essential", Submodule: "language", Action: "search"}

type SearchLanguagesQuery dyn.SearchQuery

func (SearchLanguagesQuery) CqrsRequestType() cqrs.RequestType { return searchLanguagesQueryType }

type SearchLanguagesResultData = dyn.PagedResultData[models.Language]
type SearchLanguagesResult = dyn.OpResult[SearchLanguagesResultData]

var languageExistsQueryType = cqrs.RequestType{Module: "essential", Submodule: "language", Action: "exists"}

type LanguageExistsQuery dyn.ExistsQuery

func (LanguageExistsQuery) CqrsRequestType() cqrs.RequestType { return languageExistsQueryType }

type LanguageExistsResult = dyn.OpResult[dyn.ExistsResultData]
