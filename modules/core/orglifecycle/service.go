package orglifecycle

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrgLifecycleService is what each module implements, as `OrgLifecycleAppServiceImpl` in its
// `app/` package.
//
// Both methods return an error rather than a partial answer: IAM leaves the org in provisioning
// (or deleted) on any error, and a module that silently skipped work would mark it ready with
// defaults missing.
type OrgLifecycleService interface {
	// InitOrg writes this module's defaults for the org.
	InitOrg(ctx corectx.Context, orgId model.Id) error
	// CleanUpOrg deletes every row this module holds for the org. The org row itself is IAM's.
	CleanUpOrg(ctx corectx.Context, orgId model.Id) error
}

// NoopOrgLifecycleService is for a module that holds no org-scoped data. It still answers, so
// IAM can tell "nothing to do" from "never subscribed".
type NoopOrgLifecycleService struct{}

func (NoopOrgLifecycleService) InitOrg(corectx.Context, model.Id) error {
	return nil
}

func (NoopOrgLifecycleService) CleanUpOrg(corectx.Context, model.Id) error {
	return nil
}
