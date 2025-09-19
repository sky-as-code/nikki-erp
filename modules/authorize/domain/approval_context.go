package domain

type ApprovalContext struct {
	Request      *GrantRequest
	ManagerIds   []string
	OwnerUserIds []string
	OwnerId      *string
	ResponderId  string
	Responses    []GrantResponse
}

type ResponseState struct {
	AnyManagerResponded bool
	AnyManagerDenied    bool
	AnyManagerApproved  bool
	AnyOwnerDenied      bool
	AnyOwnerApproved    bool
}

func (ctx *ApprovalContext) IsGroupReceiver() bool {
	return *ctx.Request.ReceiverType == ReceiverTypeGroup
}

func (ctx *ApprovalContext) IsResponderManager() bool {
	for _, managerId := range ctx.ManagerIds {
		if ctx.ResponderId == managerId {
			return true
		}
	}
	return false
}

func (ctx *ApprovalContext) IsResponderOwnerUser() bool {
	for _, ownerUserId := range ctx.OwnerUserIds {
		if ctx.ResponderId == ownerUserId {
			return true
		}
	}
	return false
}

func (ctx *ApprovalContext) HasAlreadyResponded() bool {
	for _, response := range ctx.Responses {
		if response.ResponderId != nil && *response.ResponderId == ctx.ResponderId {
			return true
		}
	}
	return false
}

func (ctx *ApprovalContext) GetResponseState() ResponseState {
	state := ResponseState{}

	for _, response := range ctx.Responses {
		for _, managerId := range ctx.ManagerIds {
			if *response.ResponderId == managerId {
				state.AnyManagerResponded = true
				if !*response.IsApproved {
					state.AnyManagerDenied = true
				} else {
					state.AnyManagerApproved = true
				}
				break
			}
		}

		for _, ownerUserId := range ctx.OwnerUserIds {
			if *response.ResponderId == ownerUserId {
				if !*response.IsApproved {
					state.AnyOwnerDenied = true
				} else {
					state.AnyOwnerApproved = true
				}
				break
			}
		}
	}

	return state
}