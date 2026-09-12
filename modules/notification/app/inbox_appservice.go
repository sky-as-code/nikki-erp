package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// NewInboxApplicationService serves a signed-in person their own notifications.
//
// It is built on the recipient onion rather than the notification one, because everything it
// answers is a property of the recipient row: what is unread, what order it came in, what has been
// read. The notification is joined in for its text.
func NewInboxApplicationService(
	base composable.CrudApplicationService,
	recipients it.RecipientRepository,
	cfg config.ConfigService,
) it.InboxApplicationService {
	return &InboxApplicationServiceImpl{
		CrudApplicationService: base,
		recipients:             recipients,
		cfg:                    cfg,
	}
}

type InboxApplicationServiceImpl struct {
	composable.CrudApplicationService

	recipients it.RecipientRepository
	cfg        config.ConfigService
}

func (this *InboxApplicationServiceImpl) GetInbox(
	ctx corectx.Context, query it.InboxQuery,
) (*it.InboxResult, error) {
	scope, cErrs := this.authorizeSelf(ctx, modconstants.ActionReadInbox)
	if cErrs != nil {
		return &it.InboxResult{ClientErrors: *cErrs}, nil
	}

	query.Limit = this.boundedLimit(query.Limit)
	items, err := this.recipients.Inbox(ctx, scope.orgId, scope.userId, query)
	if err != nil {
		return nil, err
	}

	// HasData is true even for an empty page. An empty inbox is a correct answer -- most
	// people have read everything most of the time -- and ServeAction turns HasData false
	// into a 400 "not found", which would make "nothing to read" indistinguishable from a
	// broken request.
	return &it.InboxResult{
		HasData: true,
		Data: it.InboxResultData{
			Items:      items,
			NextCursor: nextCursor(items, query.Limit),
		},
	}, nil
}

func (this *InboxApplicationServiceImpl) GetUnreadCount(
	ctx corectx.Context,
) (*it.UnreadCountResult, error) {
	scope, cErrs := this.authorizeSelf(ctx, modconstants.ActionReadInbox)
	if cErrs != nil {
		return &it.UnreadCountResult{ClientErrors: *cErrs}, nil
	}

	count, err := this.recipients.UnreadCount(ctx, scope.orgId, scope.userId)
	if err != nil {
		return nil, err
	}
	return &it.UnreadCountResult{
		HasData: true,
		Data:    it.UnreadCountResultData{Count: count},
	}, nil
}

// Replay answers the stream's catch-up read (BR-FS 10, BR-FS 11).
//
// It lives on the application service rather than being called straight off the repository so that
// opening a stream is authorized exactly like reading the inbox: the same person, the same
// organization, the same entitlement.
func (this *InboxApplicationServiceImpl) Replay(
	ctx corectx.Context, query it.StreamQuery, limit int,
) (*it.InboxResult, error) {
	scope, cErrs := this.authorizeSelf(ctx, modconstants.ActionReadInbox)
	if cErrs != nil {
		return &it.InboxResult{ClientErrors: *cErrs}, nil
	}

	items, err := this.recipients.Replay(ctx, scope.orgId, scope.userId, query.AfterSeq, limit)
	if err != nil {
		return nil, err
	}
	// True even when there is nothing to replay: a client that has missed nothing is the
	// normal case, not a failed lookup.
	return &it.InboxResult{
		HasData: true,
		Data:    it.InboxResultData{Items: items},
	}, nil
}

// MarkRead stamps notifications read for the calling user and nobody else.
//
// An id that belongs to someone else is not an error: the caller may legitimately hold a list that
// mixes their own notifications with ids they learned elsewhere, and refusing the whole batch would
// make one stray id lose every real read in it. Those rows simply do not match (BR 21).
func (this *InboxApplicationServiceImpl) MarkRead(
	ctx corectx.Context, command it.MarkReadCommand,
) (*it.MarkReadResult, error) {
	scope, cErrs := this.authorizeSelf(ctx, modconstants.ActionMarkRead)
	if cErrs != nil {
		return &it.MarkReadResult{ClientErrors: *cErrs}, nil
	}

	if len(command.NotificationIds) == 0 {
		errs := ft.ClientErrors{*ft.NewValidationError(
			"notification_ids", "err_notification_ids_required",
			"at least one notification id is required")}
		return &it.MarkReadResult{ClientErrors: errs}, nil
	}

	ids := make([]model.Id, 0, len(command.NotificationIds))
	for _, id := range command.NotificationIds {
		ids = append(ids, model.Id(id))
	}

	data, err := this.recipients.MarkRead(ctx, scope.orgId, scope.userId, ids)
	if err != nil {
		return nil, err
	}
	return &it.MarkReadResult{HasData: true, Data: data}, nil
}

// AuthorizeStream resolves whose stream this is, using the same authorization as every other
// inbox read so that opening a stream and listing the inbox can never disagree about who may see
// what (BR-FS 4).
func (this *InboxApplicationServiceImpl) AuthorizeStream(
	ctx corectx.Context,
) (*it.StreamScope, error) {
	scope, cErrs := this.authorizeSelf(ctx, modconstants.ActionReadInbox)
	if cErrs != nil {
		return nil, cErrs.ToError()
	}
	return &it.StreamScope{OrgId: scope.orgId, UserId: scope.userId}, nil
}

// selfScope is whose inbox a request may touch. Both values come from the request context, never
// from the payload, which is the whole of BR 18's and BR 29's isolation rule.
type selfScope struct {
	orgId  model.Id
	userId model.Id
}

func (this *InboxApplicationServiceImpl) authorizeSelf(
	ctx corectx.Context, actionCode string,
) (*selfScope, *ft.ClientErrors) {
	permissions := ctx.GetPermissions()
	userId := permissions.UserId
	if userId == "" {
		errs := ft.ClientErrors{*ft.NewUnauthenticatedError()}
		return nil, &errs
	}

	orgId, cErrs := this.resolveOrg(ctx, actionCode)
	if cErrs != nil {
		return nil, cErrs
	}
	return &selfScope{orgId: *orgId, userId: userId}, nil
}

// resolveOrg finds the one organization this read is confined to.
//
// A user may belong to several, and a notification read must name exactly one — otherwise two
// organizations' inboxes would merge into one list. The org arrives on the request like any other
// org-scoped read, and AssertAction checks the caller belongs to it.
func (this *InboxApplicationServiceImpl) resolveOrg(
	ctx corectx.Context, actionCode string,
) (*model.Id, *ft.ClientErrors) {
	params := dmodel.DynamicFields{basemodel.FieldOrgId: string(currentOrgId(ctx))}
	orgId, cErrs := this.AssertAction(ctx, actionCode, params)
	if cErrs != nil {
		return nil, cErrs
	}
	if orgId == nil {
		errs := ft.ClientErrors{*ft.NewValidationError(
			basemodel.FieldOrgId, "err_org_id_required",
			"this resource is scoped to an organization, so 'org_id' is required")}
		return nil, &errs
	}
	return orgId, nil
}

// currentOrgId reads the organization the request is already carrying, put there by the transport
// from the query string or the acting principal.
func currentOrgId(ctx corectx.Context) model.Id {
	constraints := ctx.GetDomainConstraints()
	if constraints != nil {
		if raw, found := constraints[basemodel.FieldOrgId]; found && raw != nil {
			if value, ok := raw.(string); ok {
				return model.Id(value)
			}
		}
	}

	permissions := ctx.GetPermissions()
	if permissions.Principal.OrgId != nil {
		return *permissions.Principal.OrgId
	}
	if permissions.OrgUnitOrgId != nil {
		return *permissions.OrgUnitOrgId
	}
	return ""
}

func (this *InboxApplicationServiceImpl) boundedLimit(requested int) int {
	fallback := int(this.cfg.GetInt32(modconstants.InboxDefaultPageSize, "20"))
	max := int(this.cfg.GetInt32(modconstants.InboxMaxPageSize, "100"))

	if requested <= 0 {
		return fallback
	}
	if requested > max {
		return max
	}
	return requested
}

// nextCursor is the position to resume from, and is nil on the last page.
//
// A short page means there is nothing further, so returning a cursor for it would invite a request
// that can only come back empty.
func nextCursor(items []it.InboxItem, limit int) *int64 {
	if len(items) == 0 || len(items) < limit {
		return nil
	}
	last := items[len(items)-1].StreamSeq
	return &last
}
