package grant_response

import (
	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateGrantResponseCommand)(nil)
	util.Unused(req)
}

// START: CreateGrantResponseCommand
var createGrantResponseCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "grant_response",
	Action:    "create",
}

type CreateGrantResponseCommand struct {
	RequestId   *model.Id `json:"requestId"`
	IsApproved  *bool     `json:"isApproved"`
	Reason      *string   `json:"reason"`
	ResponderId *model.Id `json:"responderId"`
}

func (CreateGrantResponseCommand) CqrsRequestType() cqrs.RequestType {
	return createGrantResponseCommandType
}

type CreateGrantResponseResult = crud.OpResult[*domain.GrantResponse]

// END: CreateGrantResponseCommand
