package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// The resource codes are aliases of the schema names rather than string literals, because
// iam_resources.code must match the schema name byte for byte. A drifted literal here denies
// every request with nothing in the error pointing back at the seed that disagrees.
//
// These three, and these actions, are exactly what 0002002_jobscheduler_iam.sql seeds. Asserting
// an action that file does not seed produces an unconditional 403 that looks like a bug in the
// permission system rather than a missing row.
const (
	jobResource       = models.JobSchemaName
	executionResource = models.ExecutionSchemaName
	attemptResource   = models.AttemptSchemaName

	actionCreate = "create"
	actionRead   = "read"
	actionUpdate = "update"
	actionDelete = "delete"
)

// assertPermission checks one action at domain scope.
//
// Domain is the only scope available to this module: the scheduler's three tables carry neither
// tenant_id nor org_id, so there is no narrower thing for a grant to be scoped to. An org-scoped
// check here could only ever match an exact or domain grant anyway, which would look like it was
// narrowing access while doing nothing.
func assertPermission(
	ctx corectx.Context, actionCode string, resourceCode string,
) *ft.ClientErrors {
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode:   actionCode,
		ResourceCode: resourceCode,
		Scope:        reguard.ResourceScopeDomain,
	})
}
