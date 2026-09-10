package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveLineDomainService closes the move line's write surface. A move line is an
// allocation the reservation engine creates and validate stamps; a client-written line would be a
// claim on a balance the balance does not know about, leaving the two irreconcilable. Editing an
// allocation by hand needs the release-and-re-reserve flow, which the transfer's own reserve and
// unreserve actions provide.
//
// The actions are refused rather than removed, so a caller gets a 400 naming the reason instead
// of a 404 that reads as a wrong URL. The reservation engine itself writes lines through the
// repository, which this does not touch.
func NewStockMoveLineDomainService(base composable.CrudDomainService) itStock.StockMoveLineDomainService {
	return &StockMoveLineDomainServiceImpl{CrudDomainService: base}
}

type StockMoveLineDomainServiceImpl struct {
	composable.CrudDomainService
}

func (this *StockMoveLineDomainServiceImpl) Create(
	_ corectx.Context, _ itStock.CreateStockMoveLineCommand, _ ...composable.CreateOptions,
) (*itStock.CreateStockMoveLineResult, error) {
	return &itStock.CreateStockMoveLineResult{ClientErrors: *moveLineNotWritable()}, nil
}

func (this *StockMoveLineDomainServiceImpl) Update(
	_ corectx.Context, _ itStock.UpdateStockMoveLineCommand, _ ...composable.UpdateOptions,
) (*itStock.UpdateStockMoveLineResult, error) {
	return &itStock.UpdateStockMoveLineResult{ClientErrors: *moveLineNotWritable()}, nil
}

func (this *StockMoveLineDomainServiceImpl) Delete(
	_ corectx.Context, _ itStock.DeleteStockMoveLineCommand, _ ...composable.DeleteOptions,
) (*itStock.DeleteStockMoveLineResult, error) {
	return &itStock.DeleteStockMoveLineResult{ClientErrors: *moveLineNotWritable()}, nil
}

func moveLineNotWritable() *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockMoveLineSchemaName,
		"stock_move_line.not_client_writable",
		"stock move lines are written by the reservation engine; reserve, unreserve or validate the "+
			"transfer instead",
	))
	return vErrs
}
