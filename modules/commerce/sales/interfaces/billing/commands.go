// Package billing declares the commands the scheduler dispatches into Sales.
//
// They arrive over the CQRS bus rather than as HTTP calls, because the scheduler is in the same
// process and a loopback request would need a routable URL and an authenticated caller for work that
// has neither.
package billing

import (
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// The compile guard: assigning each command to the interface here turns a missing CqrsRequestType
// into a build error rather than a runtime panic on the first dispatch, when the job is already
// registered and nobody is watching.
func init() {
	var _ cqrs.Request = (*IssueEinvoicesCommand)(nil)
}

var issueEinvoicesCommandType = cqrs.RequestType{
	Module: "sales", Submodule: "billing", Action: "issueEinvoices",
}

// IssueEinvoicesCommand asks Sales to issue the electronic invoices that are due.
//
// EVERY FIELD IS OPTIONAL, and that is a requirement rather than a preference: the scheduler cannot
// import this type, so it dispatches an empty JSON body. A required field would fail to unmarshal on
// every single run, and the failure would look like the job itself being broken.
//
// The fields that do exist are for a human running the job by hand — narrowing it to one
// organization while investigating, or capping how much one pass does.
type IssueEinvoicesCommand struct {
	// OrgId limits the pass to one organization. Empty means every organization, which is what the
	// scheduler sends.
	OrgId string `json:"org_id,omitempty"`

	// Limit caps how many instructions this pass issues. Zero means the job's own default.
	Limit int `json:"limit,omitempty"`
}

func (IssueEinvoicesCommand) CqrsRequestType() cqrs.RequestType {
	return issueEinvoicesCommandType
}

// IssueEinvoicesResult reports what the pass did.
//
// Counts rather than a bare success, because the scheduler records the reply and this is what an
// operator reads to tell "nothing was due" from "everything failed" — two outcomes that both leave
// no new invoices behind.
type IssueEinvoicesResult struct {
	Examined int `json:"examined"`
	Issued   int `json:"issued"`
	Failed   int `json:"failed"`

	// Indeterminate counts the attempts whose reply never came back. They are neither issued nor
	// failed, and they are the ones a human must look at: a document may exist.
	Indeterminate int `json:"indeterminate"`
}
