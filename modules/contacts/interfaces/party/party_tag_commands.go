package party

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	tag "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreatePartyTagCommand)(nil)
	req = (*UpdatePartyTagCommand)(nil)
	req = (*DeletePartyTagCommand)(nil)
	req = (*GetPartyByIdTagQuery)(nil)
	req = (*ListPartyTagsQuery)(nil)
	req = (*PartyTagExistsMultiQuery)(nil)
	util.Unused(req)
}

var createPartyTagCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "createPartyTag",
}

type CreatePartyTagCommand struct {
	Label model.LangJson `json:"label"`
}

func (CreatePartyTagCommand) CqrsRequestType() cqrs.RequestType {
	return createPartyTagCommandType
}

func (cpt CreatePartyTagCommand) ToTagCommand() tag.CreateTagCommand {
	return tag.CreateTagCommand{
		Label: cpt.Label,
	}
}

type CreatePartyTagResult = tag.CreateTagResult

var updatePartyTagCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "updatePartyTag",
}

type UpdatePartyTagCommand tag.UpdateTagCommand

func (UpdatePartyTagCommand) CqrsRequestType() cqrs.RequestType {
	return updatePartyTagCommandType
}

func (upt UpdatePartyTagCommand) ToTagCommand() tag.UpdateTagCommand {
	return tag.UpdateTagCommand(upt)
}

type UpdatePartyTagResult = tag.UpdateTagResult

var deletePartyTagCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "deletePartyTag",
}

type DeletePartyTagCommand tag.DeleteTagCommand

func (DeletePartyTagCommand) CqrsRequestType() cqrs.RequestType {
	return deletePartyTagCommandType
}

func (dpt DeletePartyTagCommand) ToTagCommand() tag.DeleteTagCommand {
	return tag.DeleteTagCommand(dpt)
}

type DeletePartyTagResult = tag.DeleteTagResult

var partyTagExistsMultiQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "partyTagExistsMulti",
}

type PartyTagExistsMultiQuery tag.TagExistsMultiQuery

type PartyTagExistsMultiResult = tag.TagExistsMultiResult

func (PartyTagExistsMultiQuery) CqrsRequestType() cqrs.RequestType {
	return partyTagExistsMultiQueryType
}

func (ptem PartyTagExistsMultiQuery) ToTagQuery() tag.TagExistsMultiQuery {
	return tag.TagExistsMultiQuery(ptem)
}

var getPartyTagByIdQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "getPartyTagById",
}

type GetPartyByIdTagQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetPartyByIdTagQuery) CqrsRequestType() cqrs.RequestType {
	return getPartyTagByIdQueryType
}

func (gptid GetPartyByIdTagQuery) ToTagQuery() tag.GetTagByIdQuery {
	return tag.GetTagByIdQuery(gptid)
}

type GetPartyTagByIdResult = tag.GetTagByIdResult

var listPartyTagsCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "party",
	Action:    "listPartyTags",
}

type ListPartyTagsQuery tag.ListTagsQuery

type ListPartyTagsResult = tag.ListTagsResult

func (ListPartyTagsQuery) CqrsRequestType() cqrs.RequestType {
	return listPartyTagsCommandType
}

func (lpt ListPartyTagsQuery) ToTagQuery() tag.ListTagsQuery {
	return tag.ListTagsQuery(lpt)
}
