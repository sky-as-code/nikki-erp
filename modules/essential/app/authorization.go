package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// Action codes for the language catalogue, seeded in `0001002_essential_iam.sql`. Reads are not
// asserted: the language list, its records and its translation bundles are what every page --
// the sign-in page included -- renders from.
const (
	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"
)

// assertLanguagePermission guards a write to the language catalogue. Languages are tenant-wide
// reference data, so the check is at tenant scope with no org.
func assertLanguagePermission(ctx corectx.Context, actionCode string) *ft.ClientErrors {
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode:   actionCode,
		ResourceCode: models.LanguageSchemaName,
		Scope:        reguard.ResourceScopeTenant,
	})
}
