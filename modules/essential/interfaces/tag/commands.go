package tag

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateTagCommand)(nil)
	req = (*DeleteTagCommand)(nil)
	req = (*GetTagQuery)(nil)
	req = (*SearchTagsQuery)(nil)
	req = (*UpdateTagCommand)(nil)
	req = (*TagExistsQuery)(nil)
	util.Unused(req)
}

var createTagCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "tag",
	Action:    "create",
}

type CreateTagCommand struct {
	models.Tag
}

func (CreateTagCommand) CqrsRequestType() cqrs.RequestType {
	return createTagCommandType
}

func (this CreateTagCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.TagSchemaName)
}

type CreateTagResult = dyn.OpResult[models.Tag]

var updateTagCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "tag",
	Action:    "update",
}

type UpdateTagCommand struct {
	models.Tag
}

func (UpdateTagCommand) CqrsRequestType() cqrs.RequestType {
	return updateTagCommandType
}

func (this UpdateTagCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.TagSchemaName)
}

type UpdateTagResult = dyn.OpResult[dyn.MutateResultData]

var deleteTagCommandType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "tag",
	Action:    "delete",
}

type DeleteTagCommand dyn.DeleteOneCommand

func (DeleteTagCommand) CqrsRequestType() cqrs.RequestType {
	return deleteTagCommandType
}

type DeleteTagResult = dyn.OpResult[dyn.MutateResultData]

var getTagQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "tag",
	Action:    "getTag",
}

type GetTagQuery dyn.GetOneQuery

func (GetTagQuery) CqrsRequestType() cqrs.RequestType {
	return getTagQueryType
}

type GetTagResult = dyn.OpResult[models.Tag]

var searchTagsQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "tag",
	Action:    "search",
}

type SearchTagsQuery dyn.SearchQuery

func (SearchTagsQuery) CqrsRequestType() cqrs.RequestType {
	return searchTagsQueryType
}

type SearchTagsResultData = dyn.PagedResultData[models.Tag]
type SearchTagsResult = dyn.OpResult[SearchTagsResultData]

var tagExistsQueryType = cqrs.RequestType{
	Module:    "essential",
	Submodule: "tag",
	Action:    "exists",
}

type TagExistsQuery dyn.ExistsQuery

func (TagExistsQuery) CqrsRequestType() cqrs.RequestType {
	return tagExistsQueryType
}

type TagExistsResult = dyn.OpResult[dyn.ExistsResultData]
