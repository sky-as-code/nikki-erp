package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

// Raising a document for another module's sale.
//
// Issue closes a draft that is already there, which suits a document someone assembled by hand. A
// caller with a sale has no draft and wants one document: the draft, its lines and the number are
// three writes that must not be separable, or a crash between them leaves a numbered document with
// no lines, or lines under a draft nobody will ever close.
//
// So all of it happens in one transaction, and the replay guard is a unique index rather than a
// check: a caller that timed out cannot know whether the document exists, and a check-then-insert
// can be raced by exactly the retry it is meant to catch.

type (
	IssueFromSourceCommand    = itInvoice.IssueFromSourceCommand
	IssueFromSourceResult     = itInvoice.IssueFromSourceResult
	IssueFromSourceResultData = itInvoice.IssueFromSourceResultData
	GetBySourceQuery          = itInvoice.GetBySourceQuery
)

// assert that the domain service really is the module's public port, so the two cannot drift apart
// without someone else's build breaking first.
var _ itInvoice.InvoiceDomainService = (*InvoiceDomainService)(nil)

// IssueFromSource raises a document for a source and issues it.
func (this *InvoiceDomainService) IssueFromSource(
	ctx corectx.Context, cmd IssueFromSourceCommand,
) (*IssueFromSourceResult, error) {
	if refusal := assertIssuableFromSource(cmd); refusal != "" {
		return refusedFromSource(refusal), nil
	}

	// Checked before the transaction as well as guarded by the index inside it. The index is what
	// makes the retry correct; this makes the common replay cheap, and answers with the original
	// document rather than a rolled-back transaction.
	existing, err := this.GetBySource(ctx, GetBySourceQuery{
		SourceType: cmd.SourceType, SourceId: cmd.SourceId,
	})
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.HasData {
		existing.Data.AlreadyExisted = true
		return existing, nil
	}

	var result *IssueFromSourceResult
	err = withInvoiceTransaction(ctx, func(tranxCtx corectx.Context) error {
		invoiceId, err := writeInvoiceDraft(tranxCtx, cmd)
		if err != nil {
			return err
		}
		if err := writeInvoiceLinesFor(tranxCtx, invoiceId, cmd); err != nil {
			return err
		}

		issued, vErrs, err := this.issueWithin(tranxCtx, invoiceId)
		if err != nil {
			return err
		}
		if vErrs != nil && vErrs.Count() > 0 {
			// A rule the caller broke, discovered only once the lines were written. The transaction
			// is rolled back so no half-built document survives.
			result = refusedFromSource(firstViolationMessage(vErrs))
			return errRollbackAfterRefusal
		}

		result = &IssueFromSourceResult{
			HasData: true,
			Data: IssueFromSourceResultData{
				InvoiceId:      issued.InvoiceId,
				Number:         issued.Number,
				IssuedAt:       issued.IssuedAt,
				SubtotalAmount: issued.SubtotalAmount,
				TaxAmount:      issued.TaxAmount,
				TotalAmount:    issued.TotalAmount,
			},
		}
		return nil
	})

	if errors.Is(err, errRollbackAfterRefusal) {
		return result, nil
	}
	if err != nil {
		// A concurrent caller won the race on the unique index. Their document is the one that
		// exists, and returning it is the whole point of the guard.
		if replayed, lookupErr := this.GetBySource(ctx, GetBySourceQuery{
			SourceType: cmd.SourceType, SourceId: cmd.SourceId,
		}); lookupErr == nil && replayed != nil && replayed.HasData {
			replayed.Data.AlreadyExisted = true
			return replayed, nil
		}
		return nil, err
	}
	return result, nil
}

// errRollbackAfterRefusal unwinds the transaction when the caller broke a rule. It never reaches
// the caller: a refusal is data, not a failure.
var errRollbackAfterRefusal = errors.New("paymentinvoice: rolling back a refused invoice")

// GetBySource reads back the document raised for one source.
func (this *InvoiceDomainService) GetBySource(
	ctx corectx.Context, query GetBySourceQuery,
) (*IssueFromSourceResult, error) {
	if query.SourceType == "" || query.SourceId == "" {
		return &IssueFromSourceResult{}, nil
	}

	engine, err := engineFor(models.InvoiceSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InvoiceFieldSourceType, dmodel.Equals, query.SourceType),
		*dmodel.NewSearchNode().NewCondition(
			models.InvoiceFieldSourceId, dmodel.Equals, query.SourceId),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetBySource")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return &IssueFromSourceResult{}, nil
	}

	invoice := models.NewInvoiceFrom(found.Data.Items[0])
	data := IssueFromSourceResultData{
		InvoiceId:      derefString(invoice.GetId()),
		Number:         derefString(invoice.GetNumber()),
		SubtotalAmount: derefDecimal(invoice.GetSubtotalAmount()),
		TaxAmount:      derefDecimal(invoice.GetTaxAmount()),
		TotalAmount:    derefDecimal(invoice.GetTotalAmount()),
	}
	if issuedAt := invoice.GetIssuedAt(); issuedAt != nil {
		data.IssuedAt = issuedAt.GoTime()
	}

	return &IssueFromSourceResult{HasData: true, Data: data}, nil
}

// assertIssuableFromSource checks what can be checked before anything is written.
func assertIssuableFromSource(cmd IssueFromSourceCommand) string {
	switch {
	case cmd.SourceType == "" || cmd.SourceId == "":
		// Without both there is no replay key, and a retry would mint a second document.
		return "a source type and id are required, because together they are what makes issuing safe to retry"
	case cmd.OrgId == "":
		return "an organization is required"
	case len(cmd.Lines) == 0:
		return "an invoice must have at least one line"
	case len(cmd.Lines) >= invoiceLinePageSize:
		return "an invoice may not carry that many lines"
	}
	return ""
}

// writeInvoiceDraft creates the draft the lines hang off.
func writeInvoiceDraft(ctx corectx.Context, cmd IssueFromSourceCommand) (string, error) {
	engine, err := engineFor(models.InvoiceSchemaName)
	if err != nil {
		return "", err
	}
	id, err := model.NewId()
	if err != nil {
		return "", err
	}

	fields := dmodel.DynamicFields{
		models.InvoiceFieldId:         string(*id),
		models.InvoiceFieldStatus:     models.InvoiceStatusDraft,
		models.InvoiceFieldSourceType: cmd.SourceType,
		models.InvoiceFieldSourceId:   cmd.SourceId,
		models.InvoiceFieldOrgId:      cmd.OrgId,
	}
	if cmd.Partner.Name != "" {
		fields[models.InvoiceFieldPartnerName] = cmd.Partner.Name
	}
	if cmd.Partner.TaxCode != "" {
		fields[models.InvoiceFieldPartnerTaxCode] = cmd.Partner.TaxCode
	}
	if cmd.Partner.Address != "" {
		fields[models.InvoiceFieldPartnerAddress] = cmd.Partner.Address
	}
	if cmd.CurrencyId != "" {
		fields[models.InvoiceFieldCurrencyId] = cmd.CurrencyId
	}
	if cmd.Note != "" {
		fields[models.InvoiceFieldNote] = cmd.Note
	}

	if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
		return "", errors.Wrap(err, "writeInvoiceDraft")
	}
	return string(*id), nil
}

// writeInvoiceLinesFor records the lines. The amounts are left for the issue step to compute: the
// quantity and the price are what a reader can check, so they are the fields that decide the total.
func writeInvoiceLinesFor(
	ctx corectx.Context, invoiceId string, cmd IssueFromSourceCommand,
) error {
	engine, err := engineFor(models.InvoiceLineSchemaName)
	if err != nil {
		return err
	}

	for _, line := range cmd.Lines {
		id, err := model.NewId()
		if err != nil {
			return err
		}
		if _, err := engine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
			models.InvoiceLineFieldId:             string(*id),
			models.InvoiceLineFieldInvoiceId:      invoiceId,
			models.InvoiceLineFieldDescription:    line.Description,
			models.InvoiceLineFieldQuantity:       line.Quantity,
			models.InvoiceLineFieldUnitPrice:      line.UnitPrice,
			models.InvoiceLineFieldTaxRatePercent: line.TaxRatePercent,
			models.InvoiceLineFieldOrgId:          cmd.OrgId,
		}); err != nil {
			return errors.Wrap(err, "writeInvoiceLinesFor")
		}
	}
	return nil
}

// refusedFromSource is a refusal the caller can act on.
func refusedFromSource(reason string) *IssueFromSourceResult {
	return &IssueFromSourceResult{Refused: true, RefusalReason: reason}
}
