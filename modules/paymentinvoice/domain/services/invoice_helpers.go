package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// withInvoiceTransaction runs body inside one transaction on a scoped copy of the context.
//
// It is the invoice's counterpart to withOrderTransaction and exists for the same reason: the
// transaction is begun on the invoice's own engine, and it goes on a clone so a committed
// transaction is not left visible to whatever runs next on the caller's context.
func withInvoiceTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	engine, err := engineFor(models.InvoiceSchemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.ResourceRepository().BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withInvoiceTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withInvoiceTransaction")
}

// findInvoiceById fetches one invoice by primary key.
func findInvoiceById(ctx corectx.Context, invoiceId string) (*models.Invoice, error) {
	engine, err := engineFor(models.InvoiceSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.InvoiceFieldId: invoiceId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findInvoiceById")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return models.NewInvoiceFrom(found.Data), nil
}

// findInvoiceLines returns every line filed against one invoice.
//
// The page is one larger than the documented maximum so the caller can tell a full invoice from an
// over-full one: reading exactly the maximum would make the two indistinguishable, and silently
// totalling only part of an invoice understates what is owed.
func findInvoiceLines(ctx corectx.Context, invoiceId string) ([]*models.InvoiceLine, error) {
	engine, err := engineFor(models.InvoiceLineSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InvoiceLineFieldInvoiceId, dmodel.Equals, invoiceId),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  invoiceLinePageSize,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findInvoiceLines")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	lines := make([]*models.InvoiceLine, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		lines = append(lines, models.NewInvoiceLineFrom(item))
	}
	return lines, nil
}

// writeInvoiceFields updates an invoice through the repository.
//
// Number, status and issued_at are declared no_update precisely so a client cannot set them; the
// issue action writes them here instead, having already checked what the transition requires.
func writeInvoiceFields(ctx corectx.Context, invoicePk string, fields dmodel.DynamicFields) error {
	engine, err := engineFor(models.InvoiceSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.InvoiceFieldId: invoicePk}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeInvoiceFields")
}

// writeInvoiceLineFields updates one line, for the same reason.
func writeInvoiceLineFields(ctx corectx.Context, linePk string, fields dmodel.DynamicFields) error {
	engine, err := engineFor(models.InvoiceLineSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.InvoiceLineFieldId: linePk}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeInvoiceLineFields")
}
