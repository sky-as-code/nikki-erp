// Package invoice is the Payment & Invoice module's port for raising a fiscal document on behalf of
// another module.
//
// It exists because the invoice engine's own Issue only closes a draft that is already there, which
// is right for a document someone assembled by hand and useless to a caller that has a sale and
// wants a document for it. Raising one is three writes that must not be separable — the draft, its
// lines, and the number — so it is one call here rather than three across a module boundary.
//
// Everything else this module does with invoices is CRUD served by the resource engine over REST. A
// caller that wants to read one should use that surface rather than reach into the domain.
package invoice

import (
	"time"

	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// InvoiceDomainService raises fiscal documents for other modules.
//
// Like the order port, there is deliberately no application-service counterpart: authorization for a
// cross-module call is established by the request that started it, so callers reach the domain
// service directly.
type InvoiceDomainService interface {
	// IssueFromSource raises a document for a source and issues it, in one transaction.
	//
	// SAFE TO RETRY. A caller that timed out cannot tell whether the document exists, so a repeat
	// call naming the same source returns the invoice the first one created, with its original
	// number and date, rather than minting a second document for one sale.
	IssueFromSource(
		ctx corectx.Context, cmd IssueFromSourceCommand,
	) (*IssueFromSourceResult, error)

	// GetBySource reads back what was raised for a source, for the caller reconciling after a
	// timeout it could not resolve at the time.
	GetBySource(ctx corectx.Context, query GetBySourceQuery) (*IssueFromSourceResult, error)
}

// IssueFromSourceCommand is one document to raise, stated in the caller's terms.
type IssueFromSourceCommand struct {
	// SourceType and SourceId name what this document is being raised for. Together they are the
	// replay key, so they must identify the *document*, not the request: two attempts at the same
	// sale must carry the same pair or the guard does not hold.
	SourceType string
	SourceId   string

	OrgId string

	// Partner is the buyer's fiscal identity AS SUPPLIED. Snapshotted onto the invoice rather than
	// referenced, so a later change of company name or address cannot rewrite what a historical
	// document said.
	Partner PartnerInfo

	// CurrencyId names the currency. A plain ulid with no foreign key: currencies belong to
	// essential, and a constraint across that boundary would make the modules undeployable apart.
	CurrencyId string

	Lines []IssueFromSourceLine

	Note string
}

// PartnerInfo is who the document is made out to.
type PartnerInfo struct {
	Name    string
	TaxCode string
	Address string
}

// IssueFromSourceLine is one line of the document.
type IssueFromSourceLine struct {
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal

	// TaxRatePercent is a PERCENTAGE — 10 for 10%, not 0.1. Sales-side rates are fractions, and the
	// conversion belongs in the adapter that crosses the boundary; stating the unit here is what
	// makes a missing conversion a bug someone can see rather than a silent hundredfold error.
	TaxRatePercent decimal.Decimal
}

// IssueFromSourceResultData is the document that now exists.
type IssueFromSourceResultData struct {
	InvoiceId string
	Number    string
	IssuedAt  time.Time

	SubtotalAmount decimal.Decimal
	TaxAmount      decimal.Decimal
	TotalAmount    decimal.Decimal

	// AlreadyExisted marks the replay path, so a caller can tell the document it just raised from
	// one an earlier attempt had already raised. Both are success.
	AlreadyExisted bool
}

// IssueFromSourceResult carries a refusal the caller can act on — no lines, a source already used
// for a different document — in Refused/RefusalReason rather than as a Go error. An error means the
// request could not be processed at all.
type IssueFromSourceResult struct {
	Data    IssueFromSourceResultData
	HasData bool

	Refused       bool
	RefusalReason string
}

// GetBySourceQuery names the source to read back.
type GetBySourceQuery struct {
	SourceType string
	SourceId   string
}
