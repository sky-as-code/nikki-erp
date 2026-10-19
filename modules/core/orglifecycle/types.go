// Package orglifecycle lets every module write its defaults for a new organization and remove
// everything it holds for a deleted one.
//
// IAM owns the organization, so IAM orchestrates: it asks every installed module, over the
// command bus, to `initOrg` after an org is created and to `cleanUpOrg` before it is removed.
// Each module answers for its own tables only. The bus allows one handler per request type, so a
// command is addressed to its target module ("sales_org.initOrg") exactly as usage checks are.
package orglifecycle

import (
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

const (
	SubmoduleName    = "org"
	ActionInitOrg    = "initOrg"
	ActionCleanUpOrg = "cleanUpOrg"
)

// RequestTypeFor builds the request type of one lifecycle action addressed to one module.
func RequestTypeFor(targetModule string, action string) cqrs.RequestType {
	return cqrs.RequestType{
		Module:    targetModule,
		Submodule: SubmoduleName,
		Action:    action,
	}
}

// OrgLifecycleCommand asks one module to initialize or clean up one organization.
//
// TargetModule and Action address the command; the handler checks both on arrival so a command
// that reaches the wrong module is refused rather than executed under another module's name.
type OrgLifecycleCommand struct {
	RequestId    string `json:"request_id"`
	TargetModule string `json:"target_module"`
	Action       string `json:"action"`
	OrgId        string `json:"org_id"`
	// RequestedBy is the user who created or deleted the org. Rows a module writes are stamped
	// with it, so the defaults read as created by the person who created the org.
	RequestedBy string `json:"requested_by"`
}

func (this OrgLifecycleCommand) CqrsRequestType() cqrs.RequestType {
	return RequestTypeFor(this.TargetModule, this.Action)
}

type OrgLifecycleResult struct {
	RequestId string `json:"request_id"`
	Module    string `json:"module"`
}
