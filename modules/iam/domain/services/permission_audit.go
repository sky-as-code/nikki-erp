package services

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
)

// permissionAuditor writes the append-only trail of grants and revocations.
//
// The cache table answers "what can this person do now"; it cannot answer "when
// did they get it, and who gave it to them", which is the question an access
// review actually asks. Every mutation that triggers a cache rebuild writes a row
// here, in the same transaction, so the two can never disagree about what happened.
type permissionAuditor struct {
	historyRepo itPerm.PermissionHistoryRepository
}

// recordRoleTransition writes one row per user affected by a role being granted or
// revoked.
//
// Failures are returned, not swallowed: the write shares the mutation's
// transaction, so an audit row that cannot be written means the change itself
// should not stand. A permission change nobody can account for afterwards is worse
// than a permission change that failed.
func (this permissionAuditor) recordRoleTransition(
	ctx corectx.Context,
	effect domain.PermissionHistoryEffect,
	reason domain.PermissionHistoryReason,
	roleId model.Id,
	userIds []model.Id,
) error {
	if this.historyRepo == nil || len(userIds) == 0 {
		return nil
	}
	approverId := actorOf(ctx)

	for _, userId := range userIds {
		entry := domain.NewPermissionHistory()
		newId, err := model.NewId()
		if err != nil {
			return err
		}
		entry.SetId(newId)
		entry.SetEffect(effect)
		entry.SetReason(reason)
		entry.SetRoleId(&roleId)
		receiver := userId
		entry.SetReceiverId(&receiver)
		entry.SetApproverId(approverId)

		if _, err := this.historyRepo.Insert(ctx, *entry); err != nil {
			return err
		}
	}
	return nil
}

// recordEntitlementTransition records a change to what a role grants, rather than
// to who holds it. The receiver is left empty on purpose: the change applies to
// every current and future holder of the role, and naming a snapshot of them here
// would age badly.
func (this permissionAuditor) recordEntitlementTransition(
	ctx corectx.Context,
	effect domain.PermissionHistoryEffect,
	reason domain.PermissionHistoryReason,
	roleId *model.Id,
	entitlementId *model.Id,
	expression *string,
) error {
	if this.historyRepo == nil {
		return nil
	}
	entry := domain.NewPermissionHistory()
	newId, err := model.NewId()
	if err != nil {
		return err
	}
	entry.SetId(newId)
	entry.SetEffect(effect)
	entry.SetReason(reason)
	entry.SetRoleId(roleId)
	entry.SetEntitlementId(entitlementId)
	entry.SetEntitlementExpr(expression)
	entry.SetApproverId(actorOf(ctx))

	_, err = this.historyRepo.Insert(ctx, *entry)
	return err
}

// actorOf names whoever made the change, when the request context knows. A
// background or migration-time call has no actor, and recording none is honest.
func actorOf(ctx corectx.Context) *model.Id {
	userId := ctx.GetPermissions().UserId
	if userId == "" {
		return nil
	}
	return &userId
}
